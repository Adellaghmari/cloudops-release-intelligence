#!/usr/bin/env python3
"""Summarize a Terraform JSON plan and fail closed on destructive actions."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


ACTION_NAMES = {
    ("create",): "CREATE",
    ("update",): "UPDATE",
    ("delete",): "DELETE",
    ("delete", "create"): "REPLACE",
    ("create", "delete"): "REPLACE",
    ("read",): "READ",
    ("no-op",): "NOOP",
}

PROTECTED_TYPES = {
    "aws_apigatewayv2_api",
    "aws_cloudfront_distribution",
    "aws_cloudwatch_event_bus",
    "aws_cloudwatch_event_rule",
    "aws_dynamodb_table",
    "aws_ecr_repository",
    "aws_iam_openid_connect_provider",
    "aws_s3_bucket",
    "aws_sqs_queue",
}


def inspect_plan(plan: dict, expected_image_uri: str | None = None) -> dict:
    counts = {"create": 0, "update": 0, "delete": 0, "replace": 0}
    changes: list[dict[str, object]] = []
    blocked: list[str] = []

    for resource in plan.get("resource_changes", []):
        address = resource.get("address", "<unknown>")
        resource_type = resource.get("type", "<unknown>")
        actions = tuple(resource.get("change", {}).get("actions", []))
        action = ACTION_NAMES.get(actions)

        if action in {"READ", "NOOP"}:
            continue
        if action is None:
            blocked.append(f"{address}: unsupported actions {list(actions)}")
            continue

        protected = resource_type in PROTECTED_TYPES
        changes.append(
            {
                "action": action,
                "address": address,
                "protected": protected,
                "type": resource_type,
            }
        )
        counts[action.lower()] += 1

        # Intentional destruction requires a separately reviewed workflow change.
        # Release dispatches never accept delete or replacement actions.
        if action in {"DELETE", "REPLACE"}:
            suffix = " protected resource" if protected else ""
            blocked.append(f"{address}: {action}{suffix}")

    if expected_image_uri:
        planned_resources: dict[str, dict] = {}

        def collect_resources(module: dict) -> None:
            for resource in module.get("resources", []):
                address = resource.get("address")
                if isinstance(address, str):
                    planned_resources[address] = resource
            for child in module.get("child_modules", []):
                collect_resources(child)

        collect_resources(plan.get("planned_values", {}).get("root_module", {}))
        for address in (
            "aws_lambda_function.api[0]",
            "aws_lambda_function.worker[0]",
        ):
            resource = planned_resources.get(address)
            if resource is None:
                blocked.append(f"{address}: missing from planned values")
                continue
            image_uri = resource.get("values", {}).get("image_uri")
            if image_uri != expected_image_uri:
                blocked.append(
                    f"{address}: image_uri {image_uri!r} does not match "
                    f"candidate {expected_image_uri!r}"
                )

    return {"blocked": blocked, "changes": changes, "counts": counts}


def render_markdown(summary: dict) -> str:
    counts = summary["counts"]
    lines = [
        "## Terraform release plan",
        "",
        f"* Create: {counts['create']}",
        f"* Update: {counts['update']}",
        f"* Delete: {counts['delete']}",
        f"* Replace: {counts['replace']}",
        "",
        "### Resource changes",
        "",
    ]
    if summary["changes"]:
        for change in summary["changes"]:
            protected = " protected" if change["protected"] else ""
            lines.append(
                f"* `{change['action']}` `{change['address']}`{protected}"
            )
    else:
        lines.append("* No resource changes")

    if summary["blocked"]:
        lines.extend(["", "### Blocked actions", ""])
        lines.extend(f"* `{item}`" for item in summary["blocked"])

    return "\n".join(lines) + "\n"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("plan_json", type=Path)
    parser.add_argument("--summary-json", type=Path, required=True)
    parser.add_argument("--summary-markdown", type=Path, required=True)
    parser.add_argument("--expected-image-uri")
    args = parser.parse_args()

    with args.plan_json.open(encoding="utf-8") as handle:
        summary = inspect_plan(
            json.load(handle), expected_image_uri=args.expected_image_uri
        )

    args.summary_json.write_text(
        json.dumps(summary, indent=2, sort_keys=True) + "\n", encoding="utf-8"
    )
    args.summary_markdown.write_text(render_markdown(summary), encoding="utf-8")

    if summary["blocked"]:
        for item in summary["blocked"]:
            print(f"Blocked release plan action: {item}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
