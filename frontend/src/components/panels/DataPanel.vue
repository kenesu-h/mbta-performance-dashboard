<script setup lang="ts">
import ApexCharts from "apexcharts";
import { computed, ref } from "vue";

import dayjs from "dayjs";

import { DETAILED_VIEW } from "@/consts";
import appStore from "@/stores/app";
import dataStore from "@/stores/data";
import mapStore from "@/stores/map";
import type { Dwell, Headway, TravelTime } from "@/types";
import { DataCategory, SelectionMode } from "@/types";
import { calculateAverages, calculateCandlesticks } from "@/utils";

const chart = ref<ApexCharts | null>(null);
const detailsShown = ref(false);

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function defaultTooltip({ seriesIndex, dataPointIndex, w }: any): string {
  return `
    <div style='padding: 0.6em'>
      <div>
        <strong>${w.globals.seriesNames[seriesIndex]}</strong>
      </div>
      <div>
        ${dayjs
          .unix(w.globals.seriesX[seriesIndex][dataPointIndex])
          .format("MMM D, HH:mm")}
        to
        ${dayjs
          .unix(w.globals.seriesX[seriesIndex][dataPointIndex])
          .add(dataStore.period, "hours")
          .format("MMM D, HH:mm")}
      </div>
      <div
        class='apexcharts-tooltip-series-group-apex-charts-active'
        style='order: 1; display: flex;'
      >
        <div class='apexcharts-tooltip-y-group'>
          <span class='apexcharts-tooltip-text-y-label'>Value: </span>
          <span class='apexcharts-tooltip-text-y-value'>
            ${w.globals.series[seriesIndex][dataPointIndex]}
          </span>
        </div>
      </div>
    </div>
  `;
}

function candlestickTooltip({ seriesIndex, dataPointIndex, w }: any): string {
  return `
    <div style='padding: 0.6em;'>
      <div>
        <strong>${w.globals.seriesNames[seriesIndex]}</strong>
      </div>
      <div>
        <span>
          ${dayjs
            .unix(w.globals.seriesX[seriesIndex][dataPointIndex])
            .format("MMM D, HH:mm")}
          to
          ${dayjs
            .unix(w.globals.seriesX[seriesIndex][dataPointIndex])
            .add(dataStore.period, "hours")
            .format("MMM D, HH:mm")}
        </span>
      </div>
      <div 
        class='apexcharts-tooltip-series-group-apexcharts-active'
        style='order: 1; display: flex;'
      >
        <div class='apexcharts-tooltip-y-group'>
          <span class='apexcharts-tooltip-text-y-label'>Open: </span>
          <span class='apexcharts-tooltip-text-y-value'>
            ${w.globals.seriesCandleO[seriesIndex][dataPointIndex]}
          </span>
        </div>
      </div>
      <div 
        class='apexcharts-tooltip-series-group-apexcharts-active'
        style='order: 2; display: flex;'
      >
        <div class='apexcharts-tooltip-y-group'>
          <span class='apexcharts-tooltip-text-y-label'>High: </span>
          <span class='apexcharts-tooltip-text-y-value'>
            ${w.globals.seriesCandleH[seriesIndex][dataPointIndex]}
          </span>
        </div>
      </div>
      <div 
        class='apexcharts-tooltip-series-group-apexcharts-active'
        style='order: 3; display: flex;'
      >
        <div class='apexcharts-tooltip-y-group'>
          <span class='apexcharts-tooltip-text-y-label'>Low: </span>
          <span class='apexcharts-tooltip-text-y-value'>
            ${w.globals.seriesCandleL[seriesIndex][dataPointIndex]}
          </span>
        </div>
      </div>
      <div 
        class='apexcharts-tooltip-series-group-apexcharts-active'
        style='order: 3; display: flex;'
      >
        <div class='apexcharts-tooltip-y-group'>
          <span class='apexcharts-tooltip-text-y-label'>Close: </span>
          <span class='apexcharts-tooltip-text-y-value'>
            ${w.globals.seriesCandleC[seriesIndex][dataPointIndex]}
          </span>
        </div>
      </div>
    </div>
  `;
}

const series = computed(() => {
  switch (dataStore.selectedCategory) {
    case DataCategory.Headway:
      return [
        {
          name: DETAILED_VIEW,
          type: "candlestick",
          data: calculateCandlesticks(
            dataStore.headways,
            (headway: Headway) => headway.currentDepDt,
            (headway: Headway) => headway.headwayTimeSec,
            dataStore.period,
          ),
        },
        {
          name: "Average Headways",
          type: "line",
          data: calculateAverages(
            dataStore.headways,
            (headway: Headway) => headway.currentDepDt,
            (headway: Headway) => headway.headwayTimeSec,
            dataStore.period,
          ),
        },
        {
          name: "Average Benchmark Headways",
          type: "line",
          data: calculateAverages(
            dataStore.headways,
            (headway: Headway) => headway.currentDepDt,
            (headway: Headway) => headway.benchmarkHeadwayTimeSec,
            dataStore.period,
          ),
        },
      ];
    case DataCategory.Dwell:
      return [
        {
          name: DETAILED_VIEW,
          type: "candlestick",
          data: calculateCandlesticks(
            dataStore.dwells,
            (dwell: Dwell) => dwell.arrDt,
            (dwell: Dwell) => dwell.dwellTimeSec,
            dataStore.period,
          ),
        },
        {
          name: "Average Dwells",
          type: "line",
          data: calculateAverages(
            dataStore.dwells,
            (dwell: Dwell) => dwell.arrDt,
            (dwell: Dwell) => dwell.dwellTimeSec,
            dataStore.period,
          ),
        },
      ];
    case DataCategory.TravelTime:
      return [
        {
          name: DETAILED_VIEW,
          type: "candlestick",
          data: calculateCandlesticks(
            dataStore.travelTimes,
            (travelTime: TravelTime) => travelTime.depDt,
            (travelTime: TravelTime) => travelTime.travelTimeSec,
            dataStore.period,
          ),
        },
        {
          name: "Average Travel Times",
          type: "line",
          data: calculateAverages(
            dataStore.travelTimes,
            (travelTime: TravelTime) => travelTime.depDt,
            (travelTime: TravelTime) => travelTime.travelTimeSec,
            dataStore.period,
          ),
        },
        {
          name: "Average Benchmark Travel Times",
          type: "line",
          data: calculateAverages(
            dataStore.travelTimes,
            (travelTime: TravelTime) => travelTime.depDt,
            (travelTime: TravelTime) => travelTime.benchmarkTravelTimeSec,
            dataStore.period,
          ),
        },
      ];
    default:
      throw new Error("Invalid data category");
  }
});

const tooltips = computed(() => {
  switch (dataStore.selectedCategory) {
    case DataCategory.Headway:
      return [candlestickTooltip, defaultTooltip, defaultTooltip];
    case DataCategory.Dwell:
      return [candlestickTooltip, defaultTooltip];
    case DataCategory.TravelTime:
      return [candlestickTooltip, defaultTooltip, defaultTooltip];
    default:
      throw new Error("Invalid data category");
  }
});

const totalPoints = computed(() => {
  switch (dataStore.selectedCategory) {
    case DataCategory.Headway:
      return dataStore.headways.length;
    case DataCategory.Dwell:
      return dataStore.dwells.length;
    case DataCategory.TravelTime:
      return dataStore.travelTimes.length;
    default:
      throw new Error("Invalid data category");
  }
});
</script>

<template>
  <Panel
    :header="`Data for the ${dataStore.selectedStop?.routeID || ''} Line`"
    :pt="{
      root: (_) => ({
        style: {
          width: '100%',
          display: 'flex',
          'flex-direction': 'column',
          'flex-grow': 1,
        },
      }),
      toggleableContent: (_) => ({
        style: {
          display: 'flex',
          'flex-direction': 'column',
          'flex-grow': 1,
        },
      }),
      content: (_) => ({
        style: {
          'flex-grow': 1,
        },
      }),
    }"
  >
    <div
      v-if="dataStore.loading"
      class="flex-column flex-grow justify-center text-align-center h-100"
    >
      <ProgressSpinner />
      <div>{{ dataStore.loadingMessage }}</div>
    </div>
    <div v-else class="flex-column gap h-100">
      <div
        v-if="dataStore.error"
        class="flex-column justify-center text-align-center h-100"
        :style="{ 'align-self': 'center' }"
      >
        <p>An error occurred while fetching data:</p>
        <Message severity="error" :closable="false">{{
          dataStore.errorMessage
        }}</Message>
        <Button
          label="Retry"
          @click="async () => await dataStore.fetchData()"
        />
      </div>
      <div v-else class="flex-column gap">
        <div class="flex-row flex-wrap gap">
          <Button
            label="Headways"
            size="small"
            :disabled="dataStore.selectedCategory === DataCategory.Headway"
            @click="
              async () => {
                dataStore.selectedCategory = DataCategory.Headway;
                await dataStore.fetchData();
                if (chart && !detailsShown) {
                  chart.hideSeries(DETAILED_VIEW);
                }
              }
            "
          >
            Headways
            <Badge
              v-tooltip.bottom="
                'The amount of time between train departures at a stop.'
              "
              value="?"
            />
          </Button>
          <Button
            size="small"
            :disabled="dataStore.selectedCategory === DataCategory.Dwell"
            @click="
              async () => {
                dataStore.selectedCategory = DataCategory.Dwell;
                await dataStore.fetchData();
                if (chart && !detailsShown) {
                  chart.hideSeries(DETAILED_VIEW);
                }
              }
            "
          >
            Dwell
            <Badge
              v-tooltip.bottom="
                'The amount of time a train spent stationary at a stop.'
              "
              value="?"
            />
          </Button>
          <Button
            label="Travel Times"
            size="small"
            :disabled="dataStore.selectedCategory === DataCategory.TravelTime"
            @click="
              async () => {
                dataStore.selectedCategory = DataCategory.TravelTime;
                await dataStore.fetchData();
                if (chart && !detailsShown) {
                  chart.hideSeries(DETAILED_VIEW);
                }
              }
            "
          >
            Travel Times
            <Badge
              v-tooltip.bottom="'The travel time between two stops.'"
              value="?"
            />
          </Button>
        </div>
        <ApexChart
          ref="chart"
          height="400px"
          :series="series"
          @mounted="
            (chart: ApexCharts) => {
              if (!detailsShown) {
                chart.hideSeries(DETAILED_VIEW);
              }
            }
          "
          @legendClick="
            (_: ApexCharts, seriesIndex: number, config: any) => {
              if (seriesIndex == 0) {
                // The candlestick data length is non-zero when the series is hidden
                detailsShown = !Boolean(config.config.series[0].data.length);
              }
            }
          "
          :options="{
            chart: {
              id: 'apexchart',
            },
            xaxis: {
              // Using datetime for the type breaks the reset zoom button
              type: 'numeric',
              labels: {
                formatter: (unix: number) => {
                  return dayjs.unix(unix).format('MMM D');
                },
              },
            },
            yaxis: {
              title: {
                text: 'Seconds',
              },
              decimalsInFloat: 2,
            },
            legend: {
              position: 'bottom',
            },
            stroke: {
              width: [1, 1],
            },
            tooltip: {
              shared: true,
              custom: tooltips,
            },
            theme: {
              mode: 'dark',
              palette: 'palette4',
            },
          }"
        />
        <div class="text-align-center">
          <span
            >Displaying data for the last 30 days. Total points:
            {{ totalPoints }}</span
          >
        </div>
        <div
          class="flex-row gap w-100"
          :class="{ 'flex-wrap': appStore.width <= 992 }"
        >
          <div class="flex-column gap w-100">
            <div class="flex-row gap">
              <label for="period" :style="{ 'font-weight': 'bold' }"
                >Period</label
              >
              <Badge
                v-tooltip.right="
                  'The amount of time each data point is calculated for. Must be between 4 and ' +
                  '360 hours.'
                "
                value="?"
              />
            </div>
            <div class="p-inputgroup">
              <span class="p-inputgroup-addon">
                <i class="pi pi-clock"></i>
              </span>
              <InputNumber
                :model-value="dataStore.period"
                input-id="period"
                suffix=" hours"
                :class="{
                  'p-invalid':
                    dataStore.period < 4 || dataStore.period > 15 * 24,
                }"
                @update:model-value="
                  (value: number) => {
                    if (value >= 4 && value <= 15 * 24) {
                      dataStore.period = value;
                    }
                  }
                "
              />
            </div>
          </div>
          <div class="flex-column gap w-100">
            <div class="flex-row gap">
              <label for="period" :style="{ 'font-weight': 'bold' }"
                >Destination</label
              >
              <Badge
                v-tooltip.right="
                  'We\'ll fetch travel times from the current station ' +
                  `(${
                    dataStore.selectedStop?.stop.name || ''
                  }) to the destination you choose ` +
                  'here.'
                "
                value="?"
              />
            </div>
            <div class="p-inputgroup">
              <span class="p-inputgroup-addon">
                <i class="pi pi-map"></i>
              </span>
              <InputText
                :model-value="`Destination: ${
                  dataStore.selectedDestination
                    ? dataStore.selectedDestination?.stop.name + ' Station'
                    : 'Choose one!'
                }`"
                input-id="period-input"
                disabled
              />
              <ToggleButton
                :model-value="mapStore.selectionMode === SelectionMode.Normal"
                on-icon="pi pi-map-marker"
                on-label=""
                off-icon="pi pi-times"
                off-label=""
                :style="{ 'z-index': 1001 }"
                @update:model-value="
                  async () => {
                    switch (mapStore.selectionMode) {
                      case SelectionMode.Normal:
                        appStore.blocked = true;
                        mapStore.selectionMode = SelectionMode.Destination;
                        scrollToTop();
                        break;
                      case SelectionMode.Destination:
                        appStore.blocked = false;
                        mapStore.selectionMode = SelectionMode.Normal;
                        break;
                      default:
                        console.error('Invalid selection mode');
                        return;
                    }
                  }
                "
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </Panel>
</template>
