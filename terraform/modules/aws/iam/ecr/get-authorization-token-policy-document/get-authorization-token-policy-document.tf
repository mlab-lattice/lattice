###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.get_authorization_token.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "get_authorization_token" {
  statement {
    effect = "Allow"

    actions = [
      "ecr:GetAuthorizationToken",
    ]

    resources = [
      "*",
    ]
  }
}
