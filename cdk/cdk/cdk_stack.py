import datetime

from aws_cdk import (
    Stack,
    aws_ec2 as ec2,
    aws_ecs as ecs,
    aws_iam as iam,
    aws_ssm as ssm,
    aws_ecr as ecr,
    aws_elasticloadbalancingv2 as elbv2,
    aws_route53 as route53,
    aws_route53_targets as targets,
    aws_certificatemanager as acm,
    CfnOutput,
)
from constructs import Construct


class CdkStack(Stack):
    def __init__(self, scope: Construct, construct_id: str, **kwargs):
        super().__init__(scope, construct_id, **kwargs)

        domain_name = "mykyyta.link"
        subdomain = f"weather-api.{domain_name}"

        # Existing ECR repository
        repo = ecr.Repository.from_repository_name(self, "WeatherApiRepo", "weather-api")

        # VPC setup
        vpc = ec2.Vpc(self, "WeatherApiVpc",
            max_azs=2,
            nat_gateways=0,
            subnet_configuration=[
                ec2.SubnetConfiguration(name="public", subnet_type=ec2.SubnetType.PUBLIC)
            ]
        )

        # ECS Cluster
        cluster = ecs.Cluster(self, "WeatherApiCluster", vpc=vpc)

        # IAM Role for ECS Task (Application Role)
        task_role = iam.Role(self, "TaskAppRole",
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("AmazonSSMReadOnlyAccess")
            ]
        )

        # IAM Role for ECS Execution
        execution_role = iam.Role(self, "TaskExecutionRole",
            assumed_by=iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
            managed_policies=[
                iam.ManagedPolicy.from_aws_managed_policy_name("service-role/AmazonECSTaskExecutionRolePolicy")
            ]
        )

        execution_role.add_to_policy(iam.PolicyStatement(
            actions=[
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:BatchGetImage"
            ],
            resources=["*"]
        ))

        # ECS Task Definition
        task_def = ecs.FargateTaskDefinition(
            self, "WeatherApiTask",
            memory_limit_mib=512,
            cpu=256,
            task_role=task_role,
            execution_role=execution_role
        )

        # SSM Secrets
        secret_keys = [
            "PORT", "DB_TYPE", "SENDGRID_API_KEY", "EMAIL_FROM",
            "WEATHER_API_KEY", "DB_URL", "JWT_SECRET", "GIN_MODE", "BASE_URL"
        ]
        secrets = {
            key: ecs.Secret.from_ssm_parameter(
                ssm.StringParameter.from_secure_string_parameter_attributes(
                    self, f"Param{key}",
                    parameter_name=f"/weather-api/{key}",
                    version=1
                )
            ) for key in secret_keys
        }

        # ECS Container Definition
        container = task_def.add_container("WeatherAppContainer",
            image=ecs.ContainerImage.from_registry(f"{repo.repository_uri}:latest"),
            logging=ecs.LogDriver.aws_logs(stream_prefix="weather"),
            secrets=secrets,
            environment={}
        )
        container.add_port_mappings(ecs.PortMapping(container_port=8080))

        # Security Groups
        alb_sg = ec2.SecurityGroup(self, "ALBSecurityGroup", vpc=vpc, allow_all_outbound=True)
        alb_sg.add_ingress_rule(ec2.Peer.any_ipv4(), ec2.Port.tcp(80), "Allow HTTP")
        alb_sg.add_ingress_rule(ec2.Peer.any_ipv4(), ec2.Port.tcp(443), "Allow HTTPS")

        ecs_sg = ec2.SecurityGroup(self, "EcsSecurityGroup", vpc=vpc, allow_all_outbound=True)
        ecs_sg.add_ingress_rule(alb_sg, ec2.Port.tcp(8080), "Allow ALB to ECS")

        # ALB
        lb = elbv2.ApplicationLoadBalancer(self, "WeatherApiLB",
            vpc=vpc,
            internet_facing=True,
            security_group=alb_sg
        )

        # Hosted Zone for domain
        hosted_zone = route53.HostedZone.from_lookup(self, "HostedZone", domain_name=domain_name)

        # ACM Certificate for HTTPS
        cert = acm.Certificate(self, "WeatherApiCert",
            domain_name=subdomain,
            validation=acm.CertificateValidation.from_dns(hosted_zone)
        )

        # HTTPS Listener
        https_listener = lb.add_listener("HTTPSListener", port=443, certificates=[cert], open=True)

        # ECS Service
        service = ecs.FargateService(self, "WeatherApiService",
            cluster=cluster,
            task_definition=task_def,
            desired_count=1,
            assign_public_ip=True,
            vpc_subnets=ec2.SubnetSelection(subnet_type=ec2.SubnetType.PUBLIC),
            security_groups=[ecs_sg]
        )

        # Register ECS Service as target
        https_listener.add_targets("WeatherApiTargets",
            port=8080,
            targets=[service],
            health_check=elbv2.HealthCheck(
                path="/health",
                port="8080",
                healthy_http_codes="200"
            )
        )

        # Route 53 Alias Record for subdomain
        route53.ARecord(self, "WeatherApiAliasRecord",
            zone=hosted_zone,
            record_name="weather-api",
            target=route53.RecordTarget.from_alias(targets.LoadBalancerTarget(lb))
        )

        # Outputs
        CfnOutput(self, "LoadBalancerURL",
            value=f"https://{subdomain}",
            description="Subdomain with HTTPS"
        )

        CfnOutput(self, "WeatherApiEcrUri",
            value=repo.repository_uri,
            description="Push image here"
        )

        CfnOutput(self, "EcsClusterName",
            value=cluster.cluster_name,
            description="ECS Cluster Name"
        )

        CfnOutput(self, "EcsServiceName",
            value=service.service_name,
            description="ECS Service Name"
        )
