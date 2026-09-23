# Smart groups

[简体中文](smart-groups.zh-CN.md)

Smart groups are **independent dynamic views**, not replacements for source groups. A node keeps its original `Group`, and the same node can match multiple smart groups. They use the node's stored **landing country** (`LinkCountry`, two-letter codes such as `GB,FR,DE`), across source groups by default. Optionally narrow the candidates by a case-insensitive node-name keyword and one or more existing source groups.

## Use

1. In **Subscriptions → Smart Groups**, create a group, choose countries from the full searchable region catalog (including PH), optionally select a node-name keyword and source groups, and set a maximum latency, minimum speed, and freshness window (default: 72 hours). Leave source groups empty to search across all original groups; new groups default to no latency ceiling. Zero for a numeric threshold means no upper/lower threshold; zero freshness disables expiry.
2. Choose an existing **node check profile** and click **Check candidates**. This starts an asynchronous test of *all matching* nodes (country, keyword, source groups), including failed or untested nodes. A TCP profile is sufficient when minimum speed is zero; choose a mihomo download-speed profile when minimum speed is positive. Refresh the member list after the task completes.
3. A node becomes a member when its latency check succeeded with a positive delay, satisfies the optional latency ceiling, and (unless freshness is zero) has a recent latency timestamp. If minimum speed is positive, the speed check must also have succeeded, meet the minimum and have a recent timestamp. Membership is evaluated on each read. Stale or failed nodes leave the view automatically.
4. In a subscription, select the smart group in **Dynamic** or **Mixed** selection mode. The subscription resolves current members when generated; its existing filtering and effective-name deduplication still apply. Direct/group/airport items precede smart-group members in the output. Preview reflects the same member criteria.

Deleting a smart group used by a subscription is rejected until it is removed from that subscription. Smart groups do not schedule their own tests: run checks manually or use the existing node-check scheduler to keep results fresh (an all-node scheduled profile also covers these countries). A manually selected check profile is used for its test settings, not its saved group/tag scope.

The member endpoint returns `candidateCount` (before health filtering), all healthy member IDs and a first-100-node display sample. If candidates are zero, check the stored `LinkCountry` on nodes and the keyword/source-group filters; if only members are zero, check latency status, thresholds, and freshness. The country chooser lists regions even if no node currently has that code. Batch-filling country from names updates the node cache immediately. No physical node group is created or moved.
