###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.autoscaling_full_access.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "autoscaling_full_access" {
  statement {
    effect = "Allow"
    actions = [
      "autoscaling:*",
    ]
    resources = [
      "*",
    ]
  }
}
