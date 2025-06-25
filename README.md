# 🏎️ PodiumPe Fantasy F1 Pricing Engine
*A transparent, data-rich salary algorithm – no machine-learning “black box” required.*

---

## 1 What the engine does
| Stage | Description | Output |
|-------|-------------|--------|
| **Ingest** | Accepts **raw driver+team JSON** (no telemetry needed). | `F1BasicDriverDataV2` & `F1TeamDataV2` |
| **Enrich** | Converts to `F1CompleteDriverV2` and fills **37 ratios** covering: 3-year pedigree, live-form, team context, driver DNA, champ %. | Every ratio stored as **raw** + **Z-score** |
| **Score** | Builds a **linear RAW score** from weighted Zs, then compresses with a logistic curve to **Strength 0-1**. | `Strength` |
| **Solve band** | Analytic solver chooses dynamic **pMin/pMax** that respect budget & roster rules (40 % – 135 % of slot). | `pMin`, `pMax` |
| **Price** | Linear map → charm-rounded price; week-to-week **elasticity** dampens shocks. | `driver.Price` |
