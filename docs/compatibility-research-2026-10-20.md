# Compatibility research — 2026-10-20

This refresh audits every YAML definition in `client/compatibility` against its linked documentation and current official release/support evidence. Each section identifies the latest released tool version found, the Kubernetes 1.35/1.36 evidence, and limitations. Prereleases and Helm chart versions are not substituted for application versions.

## Interpretation and scope

- **Published matrix** means the vendor explicitly lists the version window. **CI target** means a release-tagged workflow selects that Kubernetes version; it is not an audited successful run. **Minimum-policy** means upstream publishes a lower bound without a ceiling; it is not a version-specific test result.
- These are minor-level edges. Unless stated otherwise, release-line evidence refers to the latest applicable patch, not a guarantee that every patch in that line had identical support.
- Missing evidence is not a negative compatibility finding. The current checker nevertheless returns `NotCompatible` when an existing positive edge excludes the destination. That pre-existing engine limitation is unchanged; regression tests document current behavior, not a claim that every untested version is broken.
- Falco and Jaeger runtime matrices could not be established from applicable upstream evidence. Their release inventories are recorded in YAML comments; empty edges preserve `Unknown` rather than applying a different component’s matrix.
- OpenTelemetry Collector discovery is retained, with unknown runtime compatibility. A separate `opentelemetry-operator` node receives the Operator matrix and cert-manager relationships.
- Flux distribution versions are not controller image versions. Its node remains metadata-only; automatic distribution discovery is not implemented. Exact Kubernetes patch minima are retained. The checker now evaluates the full destination semver, with regression tests for accepted later patches and rejected .0 releases.
- Historical compatibility is not a maintenance/security-support promise. No actual Kubernetes 1.35/1.36 deployment testing is claimed by this research.

## File-by-file findings

### `argocd.yaml`

Latest released version found: **v3.5.5**.[1][2][3]

- **Kubernetes 1.35:** Tested: Argo CD 3.3, 3.4, 3.5.[1][2][3]
- **Kubernetes 1.36:** Tested: Argo CD 3.5 only.[1][2][3]

Changes:
- Added stable minors 3.4/3.5; excluded 3.6 release candidates.
- Corrected 3.3 window from 1.31-1.34 to 1.32-1.35. Revalidated all historical rows; quoted versions and used exclusive next-minor bounds.

Limitations:
- Tested-version evidence does not prove incompatibility outside these ranges. Existing coverage floor 2.10 retained.
- Checker currently labels any out-of-range combination NotCompatible even when upstream merely lacks testing. No compatible:false edges added.

Evidence: [1][2][3]
Evidence (continued): [4][5][6]
Evidence (continued): [7][8][9]
Evidence (continued): [10][11][12]
Evidence (continued): [13]

### `argoworkflows.yaml`

Latest released version found: **v4.1.5**.[14][15][16]

- **Kubernetes 1.35:** Undocumented/untested in release-4.1 tested list; no interpolation between 1.34 and 1.36.[14][15][16]
- **Kubernetes 1.36:** 4.1 explicitly tests v1.36.2; encoded 1.36 minor. Older retained minors do not document 1.36.[14][15][16]

Changes:
- Added stable minor 4.1.
- Replaced interpolated 3.6/3.7/4.0 ranges with OR constraints for explicitly tested endpoint minors; preserved all three supported 3.5 minors. Quoted versions and used exclusive bounds.

Limitations:
- CI lists exact patches (4.1: 1.34.9 and 1.36.2); graph models minor-level compatibility, not exhaustive patch testing.
- Unlisted minors are not explicitly incompatible, despite checker out-of-range behavior. Coverage floor 3.5 retained.

Evidence: [14][15][16]
Evidence (continued): [17][18][19]
Evidence (continued): [20]

### `aws-lb-controller.yaml`

Latest released version found: **v3.6.0**.[21][22]

- **Kubernetes 1.35:** Covered by official minimum-only 1.19+/1.22+ policy for 2.4+; no version-specific test confirmation.[21][22]
- **Kubernetes 1.36:** Covered by official minimum-only 1.19+/1.22+ policy for 2.4+; no version-specific test confirmation.[21][22]

Changes:
- Added released application minors 3.1-3.6; latest stable v3.6.0.
- Removed fabricated <=1.34 ceilings on 2.4+ in favor of official minimum-only policy.
- Changed historical 1.21 ceilings to <1.22, preserving Kubernetes patch support.
- Quoted every sourceVersion, including 2.10.

Limitations:
- Open policy ranges do not establish tested combinations for every historical controller patch.
- AWS 2.4.7 release notes separately require Kubernetes 1.21+ for EndpointSlice functionality.

Evidence: [21][22]

### `cert-manager.yaml`

Latest released version found: **v1.21.2**.[23][24]

- **Kubernetes 1.35:** 1.19, 1.20, 1.21[23][24]
- **Kubernetes 1.36:** 1.21[23][24]

Changes:
- Added missing released minors 1.13, 1.14, 1.20 and 1.21; replaced duplicate 1.19 typo with the published 1.20 row.
- Corrected 1.16 upper Kubernetes minor from 1.33 to 1.32 and 1.15 from 1.31 to 1.32.
- Quoted every sourceVersion; normalized all inclusive Kubernetes upper minors to exclusive next-minor bounds for full patch coverage; preserved nodes/resources/schema.

Limitations:
- 1.22 is listed as upcoming, so excluded.
- Older compatibility rows include EOL releases; compatibility is not current maintenance support.
- No published Kubernetes 1.35/1.36 evidence for retained minors outside the explicitly recorded windows.

Evidence: [23][24]

### `cilium.yaml`

Latest released version found: **v1.20.3**.[25][26][27]

- **Kubernetes 1.35:** 1.19, 1.20[25][26][27]
- **Kubernetes 1.36:** 1.20[25][26][27]

Changes:
- Added released minor 1.20.
- Replaced inaccurate historical windows with version-specific e2e matrices: 1.12=1.16-1.24, 1.13=1.16-1.26, 1.14=1.19-1.27, 1.15=1.26-1.29, 1.16=1.27-1.30, 1.17=1.29-1.32, 1.18=1.30-1.33.
- Quoted every sourceVersion; normalized all inclusive Kubernetes upper minors to exclusive next-minor bounds for full patch coverage; preserved nodes/resources/schema.

Limitations:
- Historical v1.12 rendered URL no longer exists; recovered the compatibility table from the official v1.12.19 release tag. Its generic source-rendering warning is not a development-release claim: the URL is a released tag.
- Version-specific e2e guarantees do not assert incompatibility for unlisted Kubernetes versions.
- No published 1.35 evidence before Cilium 1.19 or 1.36 evidence before 1.20.

Evidence: [25][26][27]
Evidence (continued): [28][29][30]
Evidence (continued): [31][32][33]
Evidence (continued): [34]

### `datadog.yaml`

Latest released version found: **Agent 7.84.4; Cluster Agent image 7.84.4 (2026-10-19)**.[35][36][37]

- **Kubernetes 1.35:** Agent/Cluster Agent 7.67+ satisfies the published open-ended Kubernetes 1.33+ feature-minimum policy. Not a claim that each release was tested on 1.35.[35][36][37]
- **Kubernetes 1.36:** Agent/Cluster Agent 7.67+ satisfies the published open-ended Kubernetes 1.33+ feature-minimum policy. Not a claim that each release was tested on 1.36.[35][36][37]

Changes:
- Fetched both original docUrls and the current installation/Cluster Agent documentation and latest official release.
- Added every released Agent minor from 7.19 through 7.84 (including gaps) and independently verified Cluster Agent image minors 7.39 through 7.84.
- Corrected inversion of the feature-minimum table: newer Agents are not limited to only newer Kubernetes. Within the table's Kubernetes 1.16+ scope, Agent 7.19-7.35 stops below 1.21, 7.36 below 1.22, 7.37-7.39 below 1.25, 7.40-7.66 below 1.33, and 7.67+ has no documented ceiling.
- Quoted sourceVersion strings, used exclusive next-minor upper bounds, identified minimum-only policy in comments/reasons, and preserved resource identities.
- Did not create Cluster Agent 7.37/7.38 edges: those versions appear in the documentation's threshold table but the discovered datadog/cluster-agent image has no corresponding Docker Hub tags. Removed obsolete commented pseudo-dependencies.

Limitations:
- The documentation is a feature-minimum policy, not a per-Agent-release tested Kubernetes matrix; compatibility reasons state this explicitly.
- Kubernetes 1.16 is the lower scope of the threshold table, not an asserted absolute minimum Kubernetes version. Older Kubernetes is left unmodeled.
- The documentation and Cluster Agent image tag history disagree for 7.37/7.38. Only independently verified 7.39+ image minors are added; historical Cluster Agent 1.x remains unmodeled.
- Full release and image-tag inventories are saved under k8mpatible-security-sources in the scratch directory.

Evidence: [35][36][37]
Evidence (continued): [38][39][40]
Evidence (continued): [41]

### `external-dns.yaml`

Latest released version found: **v0.23.0**.[42][43]

- **Kubernetes 1.35:** Confirmed by official >=1.33 compatibility matrix row for ExternalDNS >=0.18.x; not a claim of individual 1.35 CI execution.[42][43]
- **Kubernetes 1.36:** Confirmed by official >=1.33 compatibility matrix row for ExternalDNS >=0.18.x; not a claim of individual 1.36 CI execution.[42][43]

Changes:
- Added released application minors 0.21-0.23; latest stable v0.23.0.
- Corrected 0.18+ floor from Kubernetes 1.33 to 1.21.
- Bounded 0.10-0.17 compatibility below 1.33 and encoded the official matrix rejections.
- Replaced <=minor comparisons with <next-minor bounds.

Limitations:
- Compatibility matrix is an API compatibility policy, not a patch-by-patch testing matrix.
- Provider-specific behavior may impose additional requirements.

Evidence: [42][43]

### `external-secrets.yaml`

Latest released version found: **v2.13.0**.[44][45]

- **Kubernetes 1.35:** Official test coverage for ESO 2.0-2.8.[44][45]
- **Kubernetes 1.36:** Official test coverage for ESO 2.8-2.13.[44][45]

Changes:
- Added every released missing minor 2.1-2.13 from official support table and releases.
- Updated 2.0 to documented 1.34-1.35 coverage.
- Removed false !=minor incompatibilities: upstream table describes test coverage, not exclusive runtime compatibility.
- Converted every tested Kubernetes minor ceiling to <next minor to include patch releases.

Limitations:
- Other Kubernetes versions are undocumented/untested rather than explicitly incompatible.
- Only latest ESO minor is maintained; older entries preserve historical test coverage.

Evidence: [44][45]

### `falco.yaml`

Latest released version found: **0.45.1 (2026-10-19)**.[46][47][48]

- **Kubernetes 1.35:** Unknown for the discovered runtime image; no applicable per-runtime support matrix found. The Operator's minimum does not establish this runtime compatibility.[46][47][48]
- **Kubernetes 1.36:** Unknown for the discovered runtime image; no applicable per-runtime support matrix found. The Operator's minimum does not establish this runtime compatibility.[46][47][48]

Changes:
- Fetched the original Operator prerequisites docUrl, current runtime Helm/manifests documentation, chart metadata, and latest official Falco release.
- Replaced the operator-specific docUrl with runtime deployment documentation without changing the falcosecurity/falco DaemonSet identity.
- Removed all ten operator-derived runtime compatibility edges. Kubernetes 1.29+ is an Operator native-sidecar prerequisite, not a runtime minimum.
- Recorded verified runtime minor inventory 0.34 through 0.45 in sourced research comments, including newly released 0.44/0.45. No affirmative or negative edge is fabricated when runtime version compatibility is unknown.

Limitations:
- The existing schema has no neutral version-only/unknown edge. Released minors are inventoried in comments; edges is deliberately empty rather than representing unknown as compatible:false or an unsupported compatible:true range.
- Configured chart CI targets (including an old pinned kind node) are not a published runtime compatibility support range and do not support 1.35/1.36 claims.
- Actual runtime suitability also depends on node kernel, selected driver, plugins, privileges, and deployment configuration.

Evidence: [46][47][48]
Evidence (continued): [49][50][51]

### `flux.yaml`

Latest released version found: **v2.9.6**.[52][53][54]

- **Kubernetes 1.35:** Explicitly listed for distribution minors 2.8 and 2.9.[52][53][54]
- **Kubernetes 1.36:** Explicitly listed for distribution minor 2.9.[52][53][54]

Changes:
- Replaced all-Kubernetes placeholder with schema-conformant metadata-only flux node and distribution edges for every stable 2.x minor 2.0-2.9.
- No discovery resource invented; controller 1.x tags not conflated with Flux distribution 2.x.
- Retained exact nonzero Kubernetes patch minima; used v1.33 label conservatively for contradictory 2.8 minimum >=1.32.0.

Limitations:
- Original file had no node/docUrl/resource or version baseline. Covered stable major 2, not pre-GA 0.x.
- Metadata-only kubernetesResource: [] preserves absence of discovery; automatic Flux distribution-version discovery remains absent.
- The former checker zeroed destination patches and rejected valid 1.34.1+/1.27.1+ combinations. This PR fixes evaluation to preserve patches and adds direct Kubernetes/non-Kubernetes and shipped Flux regression tests; constraints are not broadened.
- Installation docs say 1.36 and later; bounded v2.9.0 release matrix lists only through 1.36, so no 1.37+ extrapolation.

Evidence: [52][53][54]
Evidence (continued): [55][56][57]
Evidence (continued): [58][59][60]
Evidence (continued): [61][62][63]

### `gatekeeper.yaml`

Latest released version found: **v3.23.1 (2026-08-27)**.[64][65][66]

- **Kubernetes 1.35:** Supported by upstream's explicit compatibility assumption for currently supported Gatekeeper branches 3.22/3.23. Not in their pinned CI matrix. Earlier Gatekeeper minors remain unknown on 1.35.[64][65][66]
- **Kubernetes 1.36:** Supported by upstream's explicit compatibility assumption for currently supported Gatekeeper branches 3.22/3.23. Not in their pinned CI matrix. Earlier Gatekeeper minors remain unknown on 1.36.[64][65][66]

Changes:
- Fetched original installation docUrl, latest release, tagged release-management policy, Kubernetes maintained-release list, and all v3.13.0 through v3.23.0 build_test workflow matrices.
- Added released minor 3.23; replaced invented sliding windows for all retained historical minors with the Kubernetes minor sets actually configured in tag-specific upstream CI.
- For currently supported Gatekeeper 3.22/3.23 only, added Kubernetes 1.34-1.37 under the explicit upstream support-policy assumption, including 1.34 maintenance through 2026-10-27; kept that distinct from CI targets 1.31-1.33.
- Quoted sourceVersion values and changed upper bounds to exclusive next-minor syntax. Resource discovery identity is unchanged.

Limitations:
- Support policy explicitly says assumed compatibility. It is not evidence that 1.35/1.36 was tested.
- Historical tag CI configuration identifies target minors only; no successful upstream run was audited or actual cluster runtime test performed.
- Gatekeeper’s policy links the active Kubernetes release history, which retains 1.34 in maintenance until 2026-10-27. Include that active maintenance branch for supported Gatekeeper 3.22/3.23; do not create an unsupported 1.34 hole.[164]
- No compatible:false edge was added for versions outside sourced positive ranges.

Evidence: [64][65][66]
Evidence (continued): [67][68][69]
Evidence (continued): [70][71][72]
Evidence (continued): [73][74][75]
Evidence (continued): [76][77][78]

### `istio.yaml`

Latest released version found: **1.31.2**.[79][80]

- **Kubernetes 1.35:** 1.29, 1.30, 1.31[79][80]
- **Kubernetes 1.36:** 1.30, 1.31[79][80]

Changes:
- Added released minors 1.29, 1.30 and 1.31.
- Corrected 1.26 upper Kubernetes minor from 1.32 to 1.33; used only supported, not tested-but-unsupported, matrix columns.
- Preserved the explicit released 1.31 matrix upper Kubernetes minor 1.37 (not inferred from master).
- Quoted every sourceVersion; normalized all inclusive Kubernetes upper minors to exclusive next-minor bounds for full patch coverage; preserved nodes/resources/schema.

Limitations:
- Historical EOL rows remain compatibility history, not maintenance promises.
- The official released 1.31 support row lists Kubernetes 1.37, even though this refresh primarily targets 1.35/1.36; copied the explicit matrix rather than extrapolating.
- No published 1.35 evidence before Istio 1.29 or 1.36 evidence before 1.30.

Evidence: [79][80]

### `jaeger.yaml`

Latest released version found: **v2.22.0 (2026-10-06); final v1 runtime v1.76.0**.[81][82][83]

- **Kubernetes 1.35:** Unknown for the runtime. Satisfies the current Helm chart's minimum installation requirement, but that is not a runtime/version tested support range.[81][82][83]
- **Kubernetes 1.36:** Unknown for the runtime. Satisfies the current Helm chart's minimum installation requirement, but that is not a runtime/version tested support range.[81][82][83]

Changes:
- Fetched original /docs/latest/operator/ docUrl; it returns a 404 page. Replaced it with current runtime Kubernetes deployment docs.
- Verified the discovered jaegertracing/jaeger image is the v2 runtime and latest release is 2.22.0. Recorded all released v2 minors 2.0 through 2.22 from the official changelog, including early releases paired with v1 numbers.
- Removed ten 1.52-1.62 compatibility edges copied from the retired Jaeger Operator matrix, because they describe another component and image.
- Examined current Helm chart's appVersion 2.22.0 and minimum Kubernetes 1.21-0 but did not turn an installation minimum into a runtime support/test matrix or transplant OpenTelemetry Operator compatibility. Discovery identity is unchanged.

Limitations:
- No applicable per-runtime Kubernetes support matrix found; edges is deliberately empty to represent unknown, not incompatible.
- The schema cannot inventory a tool minor without also asserting a compatibility decision, so v2 minor inventory is in research comments rather than fabricated active edges.
- The retired operator's skipped v1.58 release does not imply Jaeger runtime v1.58 was skipped: official Jaeger runtime releases include 1.58.0/1.58.1. No operator-to-runtime release mapping is made.
- Current Helm metadata is a minimum-only chart installation policy. OpenTelemetry Operator and storage/backend compatibility are separate concerns.

Evidence: [81][82][83]
Evidence (continued): [84][85][86]
Evidence (continued): [87][88]

### `keda.yaml`

Latest released version found: **v2.21.0**.[89][90]

- **Kubernetes 1.35:** 2.20, 2.21[89][90]
- **Kubernetes 1.36:** 2.21[89][90]

Changes:
- Added released minors 2.20 and 2.21 from the published tested Kubernetes matrix.
- Excluded 2.22, explicitly labeled unreleased.
- Quoted every sourceVersion; normalized all inclusive Kubernetes upper minors to exclusive next-minor bounds for full patch coverage; preserved nodes/resources/schema.

Limitations:
- The matrix describes tested versions, not proof of incompatibility outside them.
- No published 1.35 evidence before KEDA 2.20 or 1.36 evidence before 2.21.
- N-2 policy alone was not used to manufacture any range.

Evidence: [89][90]

### `kubernetes.yaml`

Latest released version found: **1.37.1 (release page)**.[67]

- **Kubernetes 1.35:** Released; upstream page lists 1.35.9.[67]
- **Kubernetes 1.36:** Released; upstream page lists 1.36.5.[67]

Changes:
- Replace n/a docUrl with upstream release history; no artificial self-compatibility edges.

Limitations:
- This YAML only identifies cluster Kubernetes version; add-on support resides in other files.

Evidence: [67]

### `kyverno.yaml`

Latest released version found: **v1.19.1**.[91][92][93]

- **Kubernetes 1.35:** 1.17, 1.18, 1.19[91][92][93]
- **Kubernetes 1.36:** unknown_no_published_matrix_evidence[91][92][93]

Changes:
- Added released minor 1.19; its published window remains 1.33-1.35 rather than inferring 1.36.
- Corrected 1.17 upper Kubernetes minor from 1.34 to 1.35 and 1.16 from 1.33 to 1.34.
- Corrected 1.11 lower Kubernetes minor from 1.26 to 1.25 and 1.10 window from 1.25-1.27 to 1.24-1.26.
- Quoted every sourceVersion; normalized all inclusive Kubernetes upper minors to exclusive next-minor bounds for full patch coverage; preserved nodes/resources/schema.

Limitations:
- No published Kubernetes 1.36 matrix support found for any released retained minor, including latest 1.19.
- Current documentation retains only the latest support window; historical rows were recovered from official release-site archives and immutable website source history.
- release-1-19-0.kyverno.io could not be fetched; official GitHub website content and an immutable historical commit supplied the missing evidence.
- Release-site names can be ahead of their table content; use the table release label (release-1-18-0 still documents v1.17), not the hostname as a version assertion.

Evidence: [91][92][93]
Evidence (continued): [94][95][96]
Evidence (continued): [97]

### `linkerd.yaml`

Latest released version found: **2.20 (numbered release, announced 2026-06-23)**.[98][99]

- **Kubernetes 1.35:** Official matrix supports Linkerd 2.19 and 2.20.[98][99]
- **Kubernetes 1.36:** Outside published numbered-release support matrix (maximum 1.35). No support inferred from rolling edge policy.[98][99]

Changes:
- Added 2.20 with official 1.31-1.35 range; extended 2.19 maximum to 1.35.
- Corrected false 2.12 maximum 1.28 to official 1.24. Revalidated historical rows; quoted versions; used exclusive next-minor bounds; preserved resource identity.

Limitations:
- Rolling edge tags (e.g. edge-26.10.1) are a distinct release stream, not missing 2.x minors. Existing coverage floor 2.12 retained.

Evidence: [98][99]

### `nginx-ingress.yaml`

Latest released version found: **5.7.0**.[100][101]

- **Kubernetes 1.35:** Tested: 5.3 through 5.7 (latest patches).[100][101]
- **Kubernetes 1.36:** Tested: 5.5 through 5.7 (latest patches).[100][101]

Changes:
- Added 5.5, 5.6, 5.7 from F5 tested-version matrix; next-minor exclusive upper bounds.

Limitations:
- Minor-only graph cannot represent support differences between patches; matrix is for latest patches per line.

Evidence: [100][101]

### `opentelemetry-collector.yaml`

Latest released version found: **v0.161.0**.[102][103][104]

- **Kubernetes 1.35:** Explicit official Operator matrix: supported for Operator minors 0.142 through 0.161 inclusive. Earlier rows were not extrapolated.[102][103][104]
- **Kubernetes 1.36:** Explicit official Operator matrix: supported for Operator 0.157-0.161; 0.157-0.159 end at 1.36, 0.160-0.161 explicitly end at 1.37. 0.156 and earlier not extended.[102][103][104]

Changes:
- Preserved original Collector node and runtime discovery with unknown compatibility; added a distinct opentelemetry-operator node. Applied all Operator matrix and cert-manager edges ONLY to the Operator.
- Fixed moved 404 docUrl to docs/getting-started/compatibility.md.
- Added all ten released Operator minors 0.152 through 0.161; all 37 released minors from 0.122 through latest now represented, excluding unreleased 0.128/0.130/0.133.
- Corrected historical Operator 0.142 upper Kubernetes minor from 1.34 to documented 1.35.
- Removed false >= Prometheus-Operator dependency edges: the official matrix column describes shipped Go/CRD packages, not a required installed Operator version.
- Restricted cert-manager v1 compatibility to >=1.0.0, <2.0.0; retained optional/default-installation dependency scope in comments.
- Quoted all sourceVersion values and encoded full-patch intervals with exclusive next-minor Kubernetes upper bounds.

Limitations:
- Filename remains opentelemetry-collector.yaml but contains separate Collector and Operator nodes. Collector 0.163.0 has no sourced runtime support matrix and no inferred Operator dependencies.
- Collector contrib latest is v0.163.0, published 2026-10-14, whereas Operator latest is v0.161.0. No Operator 0.162/0.163 edges invented from contrib releases.
- Operator and managed Collector versions can diverge with custom Spec.Image; the Operator matrix is not a standalone Collector support matrix.
- cert-manager is used for some Operator features/default installation and is not a universal Collector dependency.
- No Kubernetes cluster execution performed; evidence is upstream compatibility documentation and release artifacts.

Evidence: [102][103][104]
Evidence (continued): [105][106][107]
Evidence (continued): [108]

### `prometheus.yaml`

Latest released version found: **v0.94.1**.[109][110][111]

- **Kubernetes 1.35:** Release CI configuration explicitly selects 1.35.1 for Operator 0.90/0.91; 0.92-0.94 ranges also include 1.35 based on the documented minimum plus newer CI ceiling. Not separately CI-tested for 0.92-0.94 by these sources.[109][110][111]
- **Kubernetes 1.36:** Release CI configuration explicitly selects 1.36.1 for Operator 0.92/0.93; 0.94 selects 1.37.0 and has documented minimum 1.25.0. No 1.36 extension for 0.91 or earlier.[109][110][111]

Changes:
- Added all three missing released Operator minors 0.92, 0.93, 0.94 (latest patch 0.94.1).
- Corrected false Kubernetes minimums: >=0.84 uses >=1.25.0 by documented CEL requirement; pre-0.84 uses >=1.16.0, not the client-go minor.
- Replaced unsupported/skew-derived ceilings with conservative per-release configured kind CI ceilings: 0.76=>1.30; 0.77-0.79=>1.31; 0.80-0.82=>1.32; 0.83-0.85=>1.33; 0.86-0.89=>1.34; 0.90-0.91=>1.35; 0.92-0.93=>1.36; 0.94=>1.37.
- Retained all historical source minors, quoted sourceVersion strings, used exclusive next-minor upper bounds and documented evidence horizons versus hard support maxima.

Limitations:
- Official compatibility docs state a minimum, not a finite Kubernetes maximum. Encoded upper bounds are conservative release-CI evidence horizons, not proof that versions outside the range are incompatible. The graph currently reports out-of-range as incompatible.
- Configured CI versions are evidence of upstream test targets; this task did not inspect passing CI runs or execute cluster tests.
- Kubernetes 1.23 with CustomResourceValidationExpressions enabled is a documented conditional exception and deliberately not encoded as unconditional support.
- Older tagged compatibility pages still mention 1.16 after CEL was introduced; current upstream documentation explicitly corrects the requirement retroactively for >=0.84.
- client-go version was not interpreted as an operator minimum or as independent proof of support for newer Kubernetes minors.

Evidence: [109][110][111]
Evidence (continued): [112][113][114]
Evidence (continued): [115][116][117]
Evidence (continued): [118][119][120]
Evidence (continued): [121][122][123]
Evidence (continued): [124][125][126]
Evidence (continued): [127][128][129]
Evidence (continued): [130][131]

### `stakater-reloader.yaml`

Latest released version found: **v1.4.22**.[132][133][134]

- **Kubernetes 1.35:** Covered by upstream >=1.19 minimum-only compatibility policy; individually undocumented as tested.[132][133][134]
- **Kubernetes 1.36:** Covered by upstream >=1.19 minimum-only compatibility policy; individually undocumented as tested.[132][133][134]

Changes:
- Latest stable application remains 1.4 (v1.4.22); no missing stable minors after existing 1.3/1.4.
- Removed invented Kubernetes 1.34 ceiling, replacing it with upstream >=1.19 minimum-only policy.
- Explicitly excluded v2.0.0-beta.1 prerelease and chart-v2.x Helm versions.

Limitations:
- README announces v2 but GitHub marks application v2.0.0-beta.1 as prerelease; stable-only dataset remains 1.4.
- Minimum-only policy is not individual test confirmation.

Evidence: [132][133][134]
Evidence (continued): [135][136]

### `vault-secrets-operator.yaml`

Latest released version found: **v1.6.1**.[137][138][139]

- **Kubernetes 1.35:** Present in release-tagged VSO 1.2-1.6 CI matrices; current official VSO docs also list 1.35.[137][138][139]
- **Kubernetes 1.36:** Present from VSO 1.4.1 CI/release notes, also 1.5/1.6; current official VSO docs list 1.36.[137][138][139]

Changes:
- Removed Vault server 1.19/1.20/1.21 incorrectly used as VSO application versions.
- Added all stable VSO application minors 0.1-0.10 and 1.0-1.6, latest v1.6.1.
- Rebuilt positive ranges from release-tagged integration CI matrices; removed fabricated minimum-version incompatibilities.
- Used <next minor bounds and documented patch-specific 1.4.1 evidence.

Limitations:
- CI configuration is evidence of upstream test coverage intent, not an independently observed successful CI run.
- Minor-level schema cannot distinguish patch-specific evidence; 1.4 edge uses latest released patch 1.4.1, not all 1.4.x patches.
- Current VSO page says IBM tests Vault and lists 1.32-1.36 without historical VSO version mapping; latest VSO 1.6 CI starts at 1.33. This docs/CI discrepancy is disclosed rather than claiming a 1.32 test result; tagged VSO CI defines the conservative envelope.
- v1.6.1 CI additionally includes 1.37.0, outside current documented 1.32-1.36 support list; did not add 1.37.
- No negative edge is inferred from an untested combination.

Evidence: [137][138][139]
Evidence (continued): [140][141][142]
Evidence (continued): [143][144][145]
Evidence (continued): [146][147][148]
Evidence (continued): [149][150][151]
Evidence (continued): [152][153][154]
Evidence (continued): [155]

### `velero.yaml`

Latest released version found: **v1.18.4**.[156][157][158]

- **Kubernetes 1.35:** Official current matrix explicitly lists Kubernetes 1.35.0 tested with Velero 1.18; encoded >=1.33.0, <1.36.0, with full patches at minor granularity.[156][157][158]
- **Kubernetes 1.36:** No explicit tested 1.36 entry in the official current matrix. Left unencoded; generic expected 1.18-latest was not treated as version-specific validation.[156][157][158]

Changes:
- Repaired 404 v1.16 compatibility docUrl to current official repository compatibility matrix.
- Latest stable remains 1.18.4: existing 1.18 minor retained; no released 1.19 minor to add.
- Preserved current and historical tested-minor envelopes rather than turning expected 1.18-latest into unverified future support.
- Corrected tested lower bounds from historical tagged matrices: Velero 1.9 Kubernetes 1.22=>1.20; 1.10 1.23=>1.22; 1.11 1.24=>1.23.
- Quoted all sourceVersion strings (including 1.10) and replaced <=minor with exclusive next-minor upper bounds.

Limitations:
- Edges represent tested MINOR envelopes, not all expected-supported Kubernetes versions. Upstream expects broader 1.18-latest compatibility for modern releases and recommends testing unlisted combinations.
- Tests name particular patch versions, while this graph is minor-granular; full patch ranges are an explicit projection required by the data model, not proof each patch was individually tested.
- Older Velero release lines are retained for historical compatibility, not presented as currently maintained releases.
- The generic latest expectation is not evidence that Kubernetes 1.36 has been tested. No 1.36 positive edge was added.

Evidence: [156][157][158]
Evidence (continued): [159][160][161]
Evidence (continued): [162][163]

## Sources

[1] https://argo-cd.readthedocs.io/en/stable/operator-manual/installation
[2] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.5/docs/operator-manual/tested-kubernetes-versions.md
[3] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.4/docs/operator-manual/tested-kubernetes-versions.md
[4] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.3/docs/operator-manual/tested-kubernetes-versions.md
[5] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.2/docs/operator-manual/tested-kubernetes-versions.md
[6] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.1/docs/operator-manual/tested-kubernetes-versions.md
[7] https://raw.githubusercontent.com/argoproj/argo-cd/release-3.0/docs/operator-manual/tested-kubernetes-versions.md
[8] https://raw.githubusercontent.com/argoproj/argo-cd/release-2.14/docs/operator-manual/tested-kubernetes-versions.md
[9] https://raw.githubusercontent.com/argoproj/argo-cd/release-2.13/docs/operator-manual/tested-kubernetes-versions.md
[10] https://raw.githubusercontent.com/argoproj/argo-cd/release-2.12/docs/operator-manual/tested-kubernetes-versions.md
[11] https://raw.githubusercontent.com/argoproj/argo-cd/release-2.11/docs/operator-manual/tested-kubernetes-versions.md
[12] https://raw.githubusercontent.com/argoproj/argo-cd/release-2.10/docs/operator-manual/tested-kubernetes-versions.md
[13] https://api.github.com/repos/argoproj/argo-cd/releases/latest
[14] https://raw.githubusercontent.com/argoproj/argo-workflows/release-4.1/docs/tested-kubernetes-versions.md
[15] https://raw.githubusercontent.com/argoproj/argo-workflows/release-4.0/docs/tested-kubernetes-versions.md
[16] https://raw.githubusercontent.com/argoproj/argo-workflows/release-3.7/docs/tested-kubernetes-versions.md
[17] https://raw.githubusercontent.com/argoproj/argo-workflows/release-3.6/docs/tested-kubernetes-versions.md
[18] https://raw.githubusercontent.com/argoproj/argo-workflows/release-3.5/docs/releases.md
[19] https://argo-workflows.readthedocs.io/en/release-3.6/releases
[20] https://api.github.com/repos/argoproj/argo-workflows/releases/latest
[21] https://kubernetes-sigs.github.io/aws-load-balancer-controller/latest/deploy/installation
[22] https://github.com/kubernetes-sigs/aws-load-balancer-controller/releases/tag/v3.6.0
[23] https://api.github.com/repos/cert-manager/cert-manager/releases/latest
[24] https://cert-manager.io/docs/releases
[25] https://api.github.com/repos/cilium/cilium/releases/latest
[26] https://raw.githubusercontent.com/cilium/cilium/v1.12.19/Documentation/concepts/kubernetes/compatibility.rst
[27] https://docs.cilium.io/en/v1.13/network/kubernetes/compatibility
[28] https://docs.cilium.io/en/v1.14/network/kubernetes/compatibility
[29] https://docs.cilium.io/en/v1.15/network/kubernetes/compatibility
[30] https://docs.cilium.io/en/v1.16/network/kubernetes/compatibility
[31] https://docs.cilium.io/en/v1.17/network/kubernetes/compatibility
[32] https://docs.cilium.io/en/v1.18/network/kubernetes/compatibility
[33] https://docs.cilium.io/en/v1.19/network/kubernetes/compatibility
[34] https://docs.cilium.io/en/v1.20/network/kubernetes/compatibility
[35] https://docs.datadoghq.com/containers/cluster_agent
[36] https://docs.datadoghq.com/containers/kubernetes/installation/?tab=datadogoperator
[37] https://docs.datadoghq.com/containers/kubernetes/installation.md
[38] https://github.com/DataDog/datadog-agent/releases/tag/7.84.4
[39] https://hub.docker.com/v2/repositories/datadog/cluster-agent/tags/7.84.4
[40] https://hub.docker.com/v2/repositories/datadog/cluster-agent/tags?page_size=100&name=7.37.
[41] https://hub.docker.com/v2/repositories/datadog/cluster-agent/tags?page_size=100&name=7.38.
[42] https://github.com/kubernetes-sigs/external-dns
[43] https://github.com/kubernetes-sigs/external-dns/releases/tag/v0.23.0
[44] https://external-secrets.io/main/introduction/stability-support
[45] https://github.com/external-secrets/external-secrets/releases/tag/v2.13.0
[46] https://falco.org/docs/setup/operator
[47] https://falco.org/docs/setup/kubernetes
[48] https://raw.githubusercontent.com/falcosecurity/charts/master/charts/falco/README.md
[49] https://github.com/falcosecurity/falco/releases/tag/0.45.1
[50] https://raw.githubusercontent.com/falcosecurity/charts/master/charts/falco/Chart.yaml
[51] https://raw.githubusercontent.com/falcosecurity/falco/0.45.1/.github/kind-config.yaml
[52] https://fluxcd.io/flux/installation
[53] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.0.0
[54] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.1.0
[55] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.2.0
[56] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.3.0
[57] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.4.0
[58] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.5.0
[59] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.6.0
[60] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.7.0
[61] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.8.0
[62] https://api.github.com/repos/fluxcd/flux2/releases/tags/v2.9.0
[63] https://api.github.com/repos/fluxcd/flux2/releases/latest
[64] https://open-policy-agent.github.io/gatekeeper/website/docs/install
[65] https://github.com/open-policy-agent/gatekeeper/releases/tag/v3.23.1
[66] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.23.1/docs/Release_Management.md
[67] https://kubernetes.io/releases
[68] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.23.0/.github/workflows/workflow.yaml
[69] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.22.0/.github/workflows/workflow.yaml
[70] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.21.0/.github/workflows/workflow.yaml
[71] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.20.0/.github/workflows/workflow.yaml
[72] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.19.0/.github/workflows/workflow.yaml
[73] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.18.0/.github/workflows/workflow.yaml
[74] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.17.0/.github/workflows/workflow.yaml
[75] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.16.0/.github/workflows/workflow.yaml
[76] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.15.0/.github/workflows/workflow.yaml
[77] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.14.0/.github/workflows/workflow.yaml
[78] https://raw.githubusercontent.com/open-policy-agent/gatekeeper/v3.13.0/.github/workflows/workflow.yaml
[79] https://api.github.com/repos/istio/istio/releases/latest
[80] https://istio.io/latest/docs/releases/supported-releases
[81] https://www.jaegertracing.io/docs/latest/operator
[82] https://github.com/jaegertracing/jaeger/releases/tag/v2.22.0
[83] https://www.jaegertracing.io/docs/latest/deployment
[84] https://raw.githubusercontent.com/jaegertracing/jaeger/v2.22.0/CHANGELOG.md
[85] https://raw.githubusercontent.com/jaegertracing/jaeger-operator/main/README.md
[86] https://raw.githubusercontent.com/jaegertracing/jaeger-operator/main/COMPATIBILITY.md
[87] https://www.jaegertracing.io/docs/latest/deployment/kubernetes
[88] https://raw.githubusercontent.com/jaegertracing/helm-charts/main/charts/jaeger/Chart.yaml
[89] https://api.github.com/repos/kedacore/keda/releases/latest
[90] https://keda.sh/docs/latest/operate/cluster
[91] https://api.github.com/repos/kyverno/kyverno/releases/latest
[92] https://release-1-12-0.kyverno.io/docs/installation
[93] https://release-1-16-0.kyverno.io/docs/installation
[94] https://release-1-17-0.kyverno.io/docs/installation/installation
[95] https://release-1-18-0.kyverno.io/docs/installation/releases
[96] https://raw.githubusercontent.com/kyverno/website/049be5c9e8d7af5cac074f2a4b7cb0603001a295/src/content/docs/docs/installation/releases.md
[97] https://kyverno.io/docs/installation/releases
[98] https://linkerd.io/2026/06/23/announcing-linkerd-2.20
[99] https://linkerd.io/2-edge/reference/k8s-versions
[100] https://docs.nginx.com/nginx-ingress-controller/technical-specifications
[101] https://docs.nginx.com/nginx-ingress-controller/changelog
[102] https://github.com/open-telemetry/opentelemetry-operator/blob/v0.161.0/docs/getting-started/compatibility.md
[103] https://github.com/open-telemetry/opentelemetry-operator/blob/v0.151.0/docs/compatibility.md
[104] https://github.com/open-telemetry/opentelemetry-operator/releases/download/v0.161.0/opentelemetry-operator.yaml
[105] https://github.com/open-telemetry/opentelemetry-operator/releases/tag/v0.161.0
[106] https://github.com/open-telemetry/opentelemetry-collector-contrib/releases/tag/v0.163.0
[107] https://github.com/open-telemetry/opentelemetry-operator/blob/main/docs/compatibility.md
[108] https://opentelemetry.io/docs/collector
[109] https://prometheus-operator.dev/docs/getting-started/compatibility
[110] https://github.com/prometheus-operator/prometheus-operator/releases/tag/v0.94.1
[111] https://github.com/prometheus-operator/prometheus-operator/blob/v0.94.1/.github/actions/setup-e2e/action.yaml
[112] https://github.com/prometheus-operator/prometheus-operator/blob/v0.94.0/.github/env
[113] https://github.com/prometheus-operator/prometheus-operator/blob/v0.93.0/.github/env
[114] https://github.com/prometheus-operator/prometheus-operator/blob/v0.92.0/.github/env
[115] https://github.com/prometheus-operator/prometheus-operator/blob/v0.91.0/.github/env
[116] https://github.com/prometheus-operator/prometheus-operator/blob/v0.90.0/.github/env
[117] https://github.com/prometheus-operator/prometheus-operator/blob/v0.89.0/.github/env
[118] https://github.com/prometheus-operator/prometheus-operator/blob/v0.88.0/.github/env
[119] https://github.com/prometheus-operator/prometheus-operator/blob/v0.87.0/.github/env
[120] https://github.com/prometheus-operator/prometheus-operator/blob/v0.86.0/.github/env
[121] https://github.com/prometheus-operator/prometheus-operator/blob/v0.85.0/.github/env
[122] https://github.com/prometheus-operator/prometheus-operator/blob/v0.84.0/.github/env
[123] https://github.com/prometheus-operator/prometheus-operator/blob/v0.83.0/.github/env
[124] https://github.com/prometheus-operator/prometheus-operator/blob/v0.82.0/.github/env
[125] https://github.com/prometheus-operator/prometheus-operator/blob/v0.81.0/.github/env
[126] https://github.com/prometheus-operator/prometheus-operator/blob/v0.80.0/.github/env
[127] https://github.com/prometheus-operator/prometheus-operator/blob/v0.79.0/.github/env
[128] https://github.com/prometheus-operator/prometheus-operator/blob/v0.78.0/.github/env
[129] https://github.com/prometheus-operator/prometheus-operator/blob/v0.77.0/.github/env
[130] https://github.com/prometheus-operator/prometheus-operator/blob/v0.76.0/.github/env
[131] https://github.com/prometheus-operator/prometheus-operator/blob/v0.94.1/.github/env
[132] https://github.com/stakater/Reloader?tab=readme-ov-file
[133] https://github.com/stakater/Reloader/releases/tag/v1.4.22
[134] https://raw.githubusercontent.com/stakater/Reloader/master/README.md
[135] https://raw.githubusercontent.com/stakater/Reloader/v1.3.0/README.md
[136] https://api.github.com/repos/stakater/Reloader/releases?per_page=100
[137] https://developer.hashicorp.com/vault/docs/deploy/kubernetes/vso
[138] https://github.com/hashicorp/vault-secrets-operator/releases/tag/v1.6.1
[139] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.1.0/.github/workflows/build.yaml
[140] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.2.0/.github/workflows/build.yaml
[141] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.3.4/.github/workflows/build.yaml
[142] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.4.3/.github/workflows/build.yaml
[143] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.5.2/.github/workflows/build.yaml
[144] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.6.0/.github/workflows/build.yaml
[145] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.7.1/.github/workflows/build.yaml
[146] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.8.1/.github/workflows/build.yaml
[147] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.9.1/.github/workflows/build.yaml
[148] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v0.10.0/.github/workflows/build.yaml
[149] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.0.1/.github/workflows/build.yaml
[150] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.1.0/.github/workflows/build.yaml
[151] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.2.0/.github/workflows/build.yaml
[152] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.3.0/.github/workflows/build.yaml
[153] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.4.1/.github/workflows/build.yaml
[154] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.5.1/.github/workflows/build.yaml
[155] https://raw.githubusercontent.com/hashicorp/vault-secrets-operator/v1.6.1/.github/workflows/build.yaml
[156] https://github.com/velero-io/velero
[157] https://github.com/velero-io/velero/releases/tag/v1.18.4
[158] https://velero.io/docs/v1.16/compatibility
[159] https://github.com/velero-io/velero/blob/v1.9.0/README.md
[160] https://github.com/velero-io/velero/blob/v1.10.0/README.md
[161] https://github.com/velero-io/velero/blob/v1.11.0/README.md
[162] https://github.com/velero-io/velero/blob/v1.12.0/README.md
[163] https://github.com/velero-io/velero/blob/v1.13.0/README.md
[164] https://kubernetes.io/releases/patch-releases
