/**
 * Core Panel Plugins Registration
 *
 * 모든 내부 패널들을 명시적으로 import하여 등록합니다.
 *
 * 🔥 이제 각 패널에서 직접 panelPluginRegistry.register()를 호출하므로
 * 여기서는 단순히 import만 하면 됩니다.
 */

// ⏱️ Time Series Charts
import '../plugins/timeSeries/panel.plugin';

// 📊 Stat Panels
import '../plugins/stat/panel.plugin';

// 🥧 Distribution Charts
import '../plugins/pie/panel.plugin';

// 📋 Tables
import '../plugins/table/panel.plugin';

// 📊 Gauge Charts (비활성화)
// import '../plugins/barGauge/panel.plugin';

// 🚨 Alert Panels (비활성화)
// import '../plugins/custom_alert/panel.plugin';

// 📊 Bar Charts (비활성화)
// import '../plugins/barChart/panel.plugin';

// 📊 Histogram (비활성화)
// import '../plugins/histogram/panel.plugin';

/**
 * Core Panel 등록 완료!
 *
 * ✅ 총 4개 내부 패널 등록됨:
 * - timeSeries
 * - stat
 * - pie
 * - table
 *
 * ❌ 비활성화된 패널:
 * - barGauge
 * - barChart
 * - histogram
 * - custom_alert
 *
 * 🎉 모든 패널이 명시적으로 등록되어 추적 가능!
 */
