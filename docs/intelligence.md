# Signal Desk intelligence pipeline

The dashboard reads one public, versioned JSON snapshot from the `signal-data` branch. The scheduled GitHub Action runs three times daily. The same job checks official Render, GitHub Actions and Cloudflare limits on Mondays. It publishes only public source material and run receipts. The data branch is separate from `main`, so collection does not trigger a Render rebuild.

`GET /api/intelligence` is a read-only Glance endpoint. It fetches the fixed public snapshot URL, caches it in memory for 30 minutes and serves a clear fallback when the branch is missing. If a fetch fails or a snapshot is over 12 hours old, it suppresses the Today brief and adds a delivery alert. Radar and Build may show old evidence, explicitly labeled. Render sleep or redeploy resets this in-memory cache; nothing important is stored on Render's ephemeral filesystem.

The snapshot contract is `schema_version: 1`, `generated_at`, `data_fresh`, `sources`, `items` (`radar`, `experiments`), `alerts`, `brief`, `docs`, `cache`, and `receipt`. Source entries record status, latest item, last success and input count. Release matches require a phrase in source-authored notes and a matching project owner. Documentation text changes request human review; they do not assert a changed allowance.

## Rollout after authorization

1. Push the code to the public repository. Run the `Build Signal Desk intelligence snapshot` workflow manually once to create `signal-data`; allow Actions to write repository contents. No Render key is needed.
2. Confirm `snapshot.json` is public, schema 1, small (expected under 200 KB), and contains no private data or tokens. Check `/api/intelligence` and the four page fragments after the Render deploy.
3. Verify the first scheduled run updates `signal-data` without deploying `main`. Watch source failures for a week; tune freshness windows only with observed publisher cadence.

At nine feeds × up to 25 entries, three runs daily make at most 27 feed requests/day plus nine documentation requests on Mondays. The collector makes sequential requests with an 18-second timeout and a 2 MB response cap per source. The Action has an eight-minute job timeout. The Glance endpoint makes at most one upstream snapshot request per 30-minute running interval. GitHub schedule delays, source blocking, changed RSS structure, noisy documentation diffs, and Render cold starts remain operational risks.
