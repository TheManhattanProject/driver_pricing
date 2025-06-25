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
### Live-window example (REC)
\[
\displaystyle   \text{REC}_{\text{raw}}=\alpha P_0+(1-\alpha)\alpha P_1+\ldots,\quad
Z_{\text{REC}}=\frac{\text{REC}_{\text{raw}}-\mu}{\sigma}\frac{m}{5}\Big|_{\lvert Z\rvert\le3}
\]

### 3-Year roll-up
\[
\text{PPR}_{3y}=\frac{0.60\,PPR_0+0.36\,PPR_1+0.216\,PPR_2}{0.60+0.36+0.216}
\]

### Linear RAW
\[
\text{RAW}=0.15+\sum w_k\,Z_k
\]

### Strength
\[
S=\frac{1}{1+e^{-\text{RAW}}}\;\; (0\!\to\!1)
\]

### Price band (analytic)
\[
\begin{cases}
\displaystyle\sum p_i=\tau\cdot\text{cap}\\[2pt]
p_{\min}=m_{\min}\cdot\text{slot}\\[2pt]
p_i=a+bS_i
\end{cases}
\Longrightarrow
\begin{aligned}
b&=\frac{\tau\text{cap}-n\,p_{\min}}{\sum S_i-nS_{\min}}\\
p_{\max}&=p_{\min}+b(S_{\max}-S_{\min})
\end{aligned}
\]
