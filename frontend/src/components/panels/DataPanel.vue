<script setup lang="ts">
import { computed } from "vue";

import dayjs from "dayjs";

import appStore from "@/stores/app";
import dataStore from "@/stores/data";
import mapStore from "@/stores/map";
import type { Dwell, Headway, TravelTime } from "@/types";
import { DataCategory, SelectionMode } from "@/types";

function scrollToTop() {
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function formatPointTooltip({ seriesIndex, dataPointIndex, w }: any): string {
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

const series = computed(() => {
  switch (dataStore.selectedCategory) {
    case DataCategory.Headway: {
      const candlesticks: { x: number; y: number[] }[] = [];
      const avgHeadways: { x: number; y: number }[] = [];
      const avgBenchmarkHeadways: { x: number; y: number }[] = [];
      if (dataStore.headways.length) {
        const startOfToday = dayjs().tz("America/New_York").startOf("day");

        const chunks: Headway[][] = [];
        {
          let i = 0;
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          dataStore.headways.forEach((headway) => {
            const headwayDate = dayjs(headway.currentDepDt);

            if (i >= chunks.length) {
              chunks.push([]);
            }

            if (
              (headwayDate.isSame(startOfRange) ||
                headwayDate.isAfter(startOfRange)) &&
              headwayDate.isBefore(endOfRange)
            ) {
              chunks[i].push(headway);
            } else {
              startOfRange = startOfRange.add(dataStore.period, "hours");
              endOfRange = endOfRange.add(dataStore.period, "hours");
              i++;
            }
          });
        }

        {
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          for (let i = 0; i < chunks.length; i++) {
            const chunk = chunks[i];

            if (chunk.length) {
              let sum = 0;
              const open = chunk[0].headwayTimeSec;
              let high = chunk[0].headwayTimeSec;
              let low = chunk[0].headwayTimeSec;
              const close = chunk[chunk.length - 1].headwayTimeSec;
              let benchmarkSum = 0;

              chunk.forEach((headway) => {
                high = high
                  ? Math.max(high, headway.headwayTimeSec)
                  : headway.headwayTimeSec;
                low = low
                  ? Math.min(low, headway.headwayTimeSec)
                  : headway.headwayTimeSec;
                sum += headway.headwayTimeSec;
                benchmarkSum += headway.benchmarkHeadwayTimeSec;
              });

              const timestamp = chunk[0].currentDepDt.unix();
              candlesticks.push({
                x: timestamp,
                y: [
                  parseFloat(open.toFixed(2)),
                  parseFloat(high.toFixed(2)),
                  parseFloat(low.toFixed(2)),
                  parseFloat(close.toFixed(2)),
                ],
              });
              avgHeadways.push({
                x: timestamp,
                y: parseFloat((sum / chunk.length).toFixed(2)),
              });
              avgBenchmarkHeadways.push({
                x: timestamp,
                y: parseFloat((benchmarkSum / chunk.length).toFixed(2)),
              });
            }

            startOfRange = startOfRange.add(dataStore.period, "hours");
            endOfRange = endOfRange.add(dataStore.period, "hours");
          }
        }
      }

      return [
        {
          name: "Headways (Candlesticks)",
          type: "candlestick",
          data: candlesticks,
        },
        {
          name: "Average Headways",
          type: "line",
          data: avgHeadways,
        },
        {
          name: "Average Benchmark Headways",
          type: "line",
          data: avgBenchmarkHeadways,
        },
      ];
    }
    case DataCategory.Dwell: {
      const candlesticks: { x: number; y: number[] }[] = [];
      const avgDwells: { x: number; y: number }[] = [];
      if (dataStore.dwells.length) {
        const startOfToday = dayjs().tz("America/New_York").startOf("day");

        const chunks: Dwell[][] = [];
        {
          let i = 0;
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          dataStore.dwells.forEach((dwell) => {
            const dwellDate = dayjs(dwell.arrDt);

            if (i >= chunks.length) {
              chunks.push([]);
            }

            if (
              (dwellDate.isSame(startOfRange) ||
                dwellDate.isAfter(startOfRange)) &&
              dwellDate.isBefore(endOfRange)
            ) {
              chunks[i].push(dwell);
            } else {
              startOfRange = startOfRange.add(dataStore.period, "hours");
              endOfRange = endOfRange.add(dataStore.period, "hours");
              i++;
            }
          });
        }

        {
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          for (let i = 0; i < chunks.length; i++) {
            const chunk = chunks[i];

            if (chunk.length) {
              const open = chunk[0].dwellTimeSec;
              let high = chunk[0].dwellTimeSec;
              let low = chunk[0].dwellTimeSec;
              const close = chunk[chunk.length - 1].dwellTimeSec;
              let sum = 0;

              chunk.forEach((dwell) => {
                sum += dwell.dwellTimeSec;
                high = high
                  ? Math.max(high, dwell.dwellTimeSec)
                  : dwell.dwellTimeSec;
                low = low
                  ? Math.min(low, dwell.dwellTimeSec)
                  : dwell.dwellTimeSec;
              });

              const timestamp = chunk[0].arrDt.unix();
              candlesticks.push({
                x: timestamp,
                y: [
                  parseFloat(open.toFixed(2)),
                  parseFloat(high.toFixed(2)),
                  parseFloat(low.toFixed(2)),
                  parseFloat(close.toFixed(2)),
                ],
              });
              avgDwells.push({
                x: timestamp,
                y: parseFloat((sum / chunk.length).toFixed(2)),
              });
            }

            startOfRange = startOfRange.add(dataStore.period, "hours");
            endOfRange = endOfRange.add(dataStore.period, "hours");
          }
        }
      }

      return [
        {
          name: "Dwells (Candlesticks)",
          type: "candlestick",
          data: candlesticks,
        },
        {
          name: "Average Dwells",
          type: "line",
          data: avgDwells,
        },
      ];
    }
    case DataCategory.TravelTime: {
      const candlesticks: { x: number; y: number[] }[] = [];
      const avgTravelTimes: { x: number; y: number }[] = [];
      const avgBenchmarkTravelTimes: { x: number; y: number }[] = [];
      if (dataStore.travelTimes.length) {
        const startOfToday = dayjs().tz("America/New_York").startOf("day");

        const chunks: TravelTime[][] = [];
        {
          let i = 0;
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          dataStore.travelTimes.forEach((travelTime) => {
            const travelTimeDate = dayjs(travelTime.depDt);

            if (i >= chunks.length) {
              chunks.push([]);
            }

            if (
              (travelTimeDate.isSame(startOfRange) ||
                travelTimeDate.isAfter(startOfRange)) &&
              travelTimeDate.isBefore(endOfRange)
            ) {
              chunks[i].push(travelTime);
            } else {
              startOfRange = startOfRange.add(dataStore.period, "hours");
              endOfRange = endOfRange.add(dataStore.period, "hours");
              i++;
            }
          });
        }

        {
          let startOfRange = startOfToday.subtract(30, "days");
          let endOfRange = startOfRange.add(dataStore.period, "hours");
          for (let i = 0; i < chunks.length; i++) {
            const chunk = chunks[i];

            if (chunk.length) {
              let sum = 0;
              const open = chunk[0].travelTimeSec;
              let high = chunk[0].travelTimeSec;
              let low = chunk[0].travelTimeSec;
              const close = chunk[chunk.length - 1].travelTimeSec;
              let benchmarkSum = 0;

              chunk.forEach((travelTime) => {
                high = high
                  ? Math.max(high, travelTime.travelTimeSec)
                  : travelTime.travelTimeSec;
                low = low
                  ? Math.min(low, travelTime.travelTimeSec)
                  : travelTime.travelTimeSec;
                sum += travelTime.travelTimeSec;
                benchmarkSum += travelTime.benchmarkTravelTimeSec;
              });

              const timestamp = chunk[0].depDt.unix();
              candlesticks.push({
                x: timestamp,
                y: [
                  parseFloat(open.toFixed(2)),
                  parseFloat(high.toFixed(2)),
                  parseFloat(low.toFixed(2)),
                  parseFloat(close.toFixed(2)),
                ],
              });
              avgTravelTimes.push({
                x: timestamp,
                y: parseFloat((sum / chunk.length).toFixed(2)),
              });
              avgBenchmarkTravelTimes.push({
                x: timestamp,
                y: parseFloat((benchmarkSum / chunk.length).toFixed(2)),
              });
            }

            startOfRange = startOfRange.add(dataStore.period, "hours");
            endOfRange = endOfRange.add(dataStore.period, "hours");
          }
        }
      }

      return [
        {
          name: "Travel Times (Candlesticks)",
          type: "candlestick",
          data: candlesticks,
        },
        {
          name: "Average Travel Times",
          type: "line",
          data: avgTravelTimes,
        },
        {
          name: "Average Benchmark Travel Times",
          type: "line",
          data: avgBenchmarkTravelTimes,
        },
      ];
    }
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
              }
            "
          >
            Headways
            <Badge
              v-tooltip.bottom="
                'The amount of time between the departures of the previous and ' +
                'current trains at a stop. The higher it is, the worse it is.'
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
          height="400px"
          :series="series"
          :options="{
            chart: {
              id: 'vuechart',
            },
            xaxis: {
              // Note: Using datetime for the type breaks the reset zoom button
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
            plotOptions: {
              candlestick: {
                colors: {
                  upward: '#EF403C',
                  downward: '#00B746',
                },
              },
            },
            tooltip: {
              shared: true,
              custom: [
                ({ seriesIndex, dataPointIndex, w }: any) => {
                  return `
                    <div style='padding: 0.6em;'>
                      <div>
                        <strong>${w.globals.seriesNames[seriesIndex]}</strong>
                      </div>
                      <div>
                        <span>
                          ${dayjs
                            .unix(
                              w.globals.seriesX[seriesIndex][dataPointIndex],
                            )
                            .format('MMM D, HH:mm')}
                          to
                          ${dayjs
                            .unix(
                              w.globals.seriesX[seriesIndex][dataPointIndex],
                            )
                            .add(dataStore.period, 'hours')
                            .format('MMM D, HH:mm')}
                        </span>
                      </div>
                      <div 
                        class='apexcharts-tooltip-series-group-apexcharts-active'
                        style='order: 1; display: flex;'
                      >
                        <div class='apexcharts-tooltip-y-group'>
                          <span class='apexcharts-tooltip-text-y-label'>Open: </span>
                          <span class='apexcharts-tooltip-text-y-value'>
                            ${
                              w.globals.seriesCandleO[seriesIndex][
                                dataPointIndex
                              ]
                            }
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
                            ${
                              w.globals.seriesCandleH[seriesIndex][
                                dataPointIndex
                              ]
                            }
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
                            ${
                              w.globals.seriesCandleL[seriesIndex][
                                dataPointIndex
                              ]
                            }
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
                            ${
                              w.globals.seriesCandleC[seriesIndex][
                                dataPointIndex
                              ]
                            }
                          </span>
                        </div>
                      </div>
                    </div>
                  `;
                },
                formatPointTooltip,
                formatPointTooltip,
              ],
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
                  'The amount of time each data point represents.'
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
                  'The dashboard will retrieve travel times from the current ' +
                  'station to whatever destination you select.'
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
