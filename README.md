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


## 2 Key maths (cheat-sheet)

### Live-window EWMA (REC)
$$
\operatorname{REC}_{\text{raw}}
    = \alpha P_0 + (1-\alpha)\alpha P_1 + (1-\alpha)^2\alpha P_2 + \dots
$$

$$
Z_{\text{REC}}
    = \frac{\operatorname{REC}_{\text{raw}}-\mu}{\sigma}\cdot\frac{m}{5},
    \qquad |Z|\le 3
$$

### 3-year roll-up (PPR)

$$
\operatorname{PPR}_{3y}
  = \frac{0.60\,\operatorname{PPR}_0 + 0.36\,\operatorname{PPR}_1 + 0.216\,\operatorname{PPR}_2}
         {0.60 + 0.36 + 0.216}
$$

### Linear RAW score

$$
\text{RAW} = 0.15 + \sum_k w_k\,Z_k
$$

### Strength mapping

$$
S = \frac{1}{1 + e^{-\text{RAW}}}\qquad (0 \;\longrightarrow\; 1)
$$

### Analytic price band

$$
\begin{aligned}
\sum_i p_i &= \tau \cdot \text{cap},\\[4pt]
p_{\min}   &= m_{\min}\,\text{slot},\\[4pt]
p_i        &= a + b\,S_i
\end{aligned}
$$

$$
\Longrightarrow\qquad
\begin{aligned}
b        &= \frac{\tau\,\text{cap} - n\,p_{\min}}{\sum_i S_i - n S_{\min}},\\[6pt]
p_{\max} &= p_{\min} + b\,(S_{\max} - S_{\min})
\end{aligned}
$$
