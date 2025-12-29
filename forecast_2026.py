#!/usr/bin/env python3
import pandas as pd
import numpy as np
import sys
from datetime import datetime

# Load real data
try:
    nq = pd.read_csv('NQ_2024_2025.csv', parse_dates=['date']).set_index('date')
    es = pd.read_csv('ES_2024_2025.csv', parse_dates=['date']).set_index('date')
    data = nq[['close']].join(es[['close']], lsuffix='_nq', rsuffix='_es').dropna()
    print(f"✅ Loaded real  {len(data)} days (2024–2025)")
except Exception as e:
    print(f"❌ Failed to load CSVs: {e}")
    sys.exit(1)

# Biweekly (10 trading days)
periods = []
dates = data.index
i = 0
p = 1
while i + 9 < len(dates):
    start = dates[i]
    end = dates[i+9]
    r_nq = (data.loc[end, 'close_nq'] / data.loc[start, 'close_nq'] - 1) * 100
    r_es = (data.loc[end, 'close_es'] / data.loc[start, 'close_es'] - 1) * 100
    spread = r_nq - r_es
    periods.append([p, start.date(), end.date(), r_nq, r_es, spread])
    i += 10
    p += 1

hist = pd.DataFrame(periods, columns=['period','start','end','nq','es','spread'])
spreads = hist['spread'].values

# Monte Carlo forecast for 26 periods of 2026
np.random.seed(2026)
bus_days = pd.bdate_range('2026-01-02', periods=260)
forecasts = []
for i in range(26):
    s, e = bus_days[i*10], bus_days[i*10+9]
    samples = np.random.choice(spreads, 10000, replace=True)
    forecasts.append({
        'period': i+1,
        'start_date': s.date(),
        'end_date': e.date(),
        'expected_spread': round(samples.mean(), 3),
        'prob_outperform': round((samples > 0).mean() * 100, 1)
    })

f = pd.DataFrame(forecasts)
cum = f['expected_spread'].sum()

# Report
report = f"""# 📊 2026 Futures Forecast — REAL DATA FROM MASSIVE.COM

✅ Data source: CME E-mini futures (NQ/ES) via Massive API  
✅ Period: 2024–2025 (real daily settlements)  
✅ Forecast: 26 biweekly periods in 2026

## Results
- **Expected cumulative NQ outperformance**: {cum:+.2f}%
- **Avg. biweekly edge**: {f['expected_spread'].mean():+.3f}%
- **High-confidence periods (P > 65%)**: {len(f[f['prob_outperform']>65])}

Generated: {datetime.utcnow().strftime('%Y-%m-%d %H:%M UTC')}
"""
print(report)
with open('FORECAST_REPORT.md', 'w') as f_out:
    f_out.write(report)

hist.to_csv('historical_spreads.csv', index=False)
f.to_csv('2026_forecast.csv', index=False)
