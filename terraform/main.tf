terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "fireline-terraform-state"
    key            = "fireline/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "fireline-terraform-locks"
    encrypt        = true
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "fireline"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# ---------- VPC ----------
module "vpc" {
  source = "./modules/vpc"

  environment     = var.environment
  vpc_cidr        = var.vpc_cidr
  azs             = var.azs
  private_subnets = var.private_subnets
  public_subnets  = var.public_subnets
}

# ---------- ECR ----------
module "ecr" {
  source      = "./modules/ecr"
  environment = var.environment
}

# ---------- RDS (PostgreSQL) ----------
module "rds" {
  source = "./modules/rds"

  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  db_instance_class  = var.db_instance_class
  db_name            = "fireline"
  db_username        = "fireline"
  ecs_security_group_id = module.ecs.ecs_security_group_id
}

# ---------- ElastiCache (Redis) ----------
module "redis" {
  source = "./modules/redis"

  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  node_type          = var.redis_node_type
  ecs_security_group_id = module.ecs.ecs_security_group_id
}

# ---------- S3 ----------
module "s3" {
  source      = "./modules/s3"
  environment = var.environment
}

# ---------- Secrets Manager ----------
module "secrets" {
  source      = "./modules/secrets"
  environment = var.environment
}

# ---------- ECS (Fargate) ----------
module "ecs" {
  source = "./modules/ecs"

  environment        = var.environment
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  public_subnet_ids  = module.vpc.public_subnet_ids
  ecr_repository_url = module.ecr.repository_url
  db_secret_arn      = module.secrets.db_secret_arn
  redis_endpoint     = module.redis.endpoint
  container_cpu      = var.container_cpu
  container_memory   = var.container_memory
  desired_count      = var.desired_count
}
