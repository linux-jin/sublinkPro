# Smart groups

[简体中文](smart-groups.zh-CN.md)

Smart groups are **independent dynamic views**, not replacements for source groups. A node keeps its original `Group`, and the same node can match multiple smart groups. They use the node's stored **landing country** (`LinkCountry`, two-letter codes such as `GB,FR,DE`), regardless of source group.

## Use

1. In **Subscriptions → Smart Groups**, create a group, choose countries and optionally a maximum latency, minimum speed, and freshness window (default: 72 hours). Zero for a numeric threshold means no upper/lower threshold; zero freshness disables expiry.
2. Choose an existing **node check profile** and click **Check candidates**. This starts an asynchronous test of *all* nodes in those countries, including failed or untested nodes. The profile should actually test both latency and download speed; a TCP-only latency profile without a successful speed result will not produce members. Refresh the member list after the task completes.
3. A node becomes a member only when **both** latency and speed have successful, positive stored results, satisfy the thresholds, and (unless freshness is zero) both timestamps are recent. Membership is evaluated on each read. Stale or failed nodes leave the view automatically.
4. In a subscription, select the smart group in **Dynamic** or **Mixed** selection mode. The subscription resolves current members when generated; its existing filtering and effective-name deduplication still apply. Direct/group/airport items precede smart-group members in the output. Preview reflects the same member criteria.

Deleting a smart group used by a subscription is rejected until it is removed from that subscription. Smart groups do not schedule their own tests: run checks manually or use the existing node-check scheduler to keep results fresh (an all-node scheduled profile also covers these countries). A manually selected check profile is used for its test settings, not its saved group/tag scope.

The member endpoint returns all member IDs and a first-100-node display sample. No physical node group is created or moved.
