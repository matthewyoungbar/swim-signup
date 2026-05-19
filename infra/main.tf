terraform {
  required_version = ">= 1.5"
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region"            { default = "us-east-1" }
variable "google_calendar_id"    { description = "Google Calendar ID to sync from" }
variable "google_credentials_json" {
  description = "Google service account JSON (keep in secrets manager in prod)"
  sensitive   = true
  default     = ""
}

# ─── DynamoDB Tables ─────────────────────────────────────────────────────────

resource "aws_dynamodb_table" "practices" {
  name         = "swim-practices"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute { name = "id"; type = "S" }

  ttl { attribute_name = "ttl"; enabled = true }

  tags = { App = "swim-signup" }
}

resource "aws_dynamodb_table" "signups" {
  name         = "swim-signups"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "practiceId"
  range_key    = "swimmerEmail"

  attribute { name = "practiceId";   type = "S" }
  attribute { name = "swimmerEmail"; type = "S" }

  # GSI so we can query all practices for a given swimmer
  global_secondary_index {
    name            = "swimmerEmail-index"
    hash_key        = "swimmerEmail"
    projection_type = "ALL"
  }

  tags = { App = "swim-signup" }
}

# ─── IAM ─────────────────────────────────────────────────────────────────────

resource "aws_iam_role" "lambda_role" {
  name = "swim-signup-lambda-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "lambda_policy" {
  role = aws_iam_role.lambda_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:DeleteItem",
          "dynamodb:UpdateItem", "dynamodb:Query", "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.practices.arn,
          aws_dynamodb_table.signups.arn,
          "${aws_dynamodb_table.signups.arn}/index/*"
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

# ─── Lambda ───────────────────────────────────────────────────────────────────

resource "aws_lambda_function" "swim_signup" {
  function_name    = "swim-signup-api"
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"         # Go Lambda uses "bootstrap" binary name
  runtime          = "provided.al2023"   # Go uses custom runtime
  filename         = "../backend/lambda.zip"
  source_code_hash = filebase64sha256("../backend/lambda.zip")
  timeout          = 30
  memory_size      = 256

  environment {
    variables = {
      GOOGLE_CALENDAR_ID      = var.google_calendar_id
      GOOGLE_CREDENTIALS_JSON = var.google_credentials_json
    }
  }

  tags = { App = "swim-signup" }
}

# ─── Lambda Function URL (no API Gateway needed) ──────────────────────────────

resource "aws_lambda_function_url" "swim_signup_url" {
  function_name      = aws_lambda_function.swim_signup.function_name
  authorization_type = "NONE"   # Public; add NONE + own auth or AWS_IAM for private

  cors {
    allow_credentials = false
    allow_origins     = ["*"]   # Restrict to your frontend domain in production
    allow_methods     = ["GET", "POST", "DELETE", "OPTIONS"]
    allow_headers     = ["Content-Type", "X-Swimmer-Email"]
    max_age           = 300
  }
}

# ─── Outputs ─────────────────────────────────────────────────────────────────

output "api_url" {
  description = "Lambda Function URL — set as VITE_API_URL in the frontend"
  value       = aws_lambda_function_url.swim_signup_url.function_url
}

output "practices_table" { value = aws_dynamodb_table.practices.name }
output "signups_table"   { value = aws_dynamodb_table.signups.name }
