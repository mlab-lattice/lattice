###############################################################################
# Output

output "json" {
  value = "${data.aws_iam_policy_document.ec2_read_only.json}"
}

###############################################################################
# IAM

data "aws_iam_policy_document" "ec2_read_only" {
  statement {
    effect = "Allow"
    actions = [
      "ec2:Describe*",
    ]
    resources = [
      "*",
    ]
  }
}
