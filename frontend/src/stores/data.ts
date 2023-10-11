import axios, { AxiosResponse } from "axios";
import dayjs from "dayjs";
import { reactive } from "vue";

import { ENV } from "@/consts";
import {
  BackendResponse,
  DataCategory,
  Dwell,
  Headway,
  LoadingMessage,
  TravelTime,
  RawDwell,
  RawHeadway,
  RawTravelTime,
  RouteID,
  RouteStop,
  Stop,
} from "@/types";
import { reduceStopIDs } from "@/utils";

interface Store {
  selectedStop: RouteStop | null;
  selectedDestination: RouteStop | null;
  selectedCategory: DataCategory;

  period: number;
  headways: Headway[];
  dwells: Dwell[];
  travelTimes: TravelTime[];

  loading: boolean;
  loadingMessage: LoadingMessage;

  error: boolean;
  errorMessage: string;

  toastMessage: string;

  selectStop: (a: Stop, b: RouteID) => Promise<void>;
  selectDestination: (a: Stop, b: RouteID) => Promise<void>;
  fetchData: () => Promise<void>;
}

const store = reactive<Store>({
  selectedStop: null,
  selectedDestination: null,
  selectedCategory: DataCategory.Headway,

  period: 16,
  headways: [],
  dwells: [],
  travelTimes: [],

  loading: false,
  loadingMessage: LoadingMessage.None,

  error: false,
  errorMessage: "",

  toastMessage: "",

  async selectStop(stop: Stop, routeID: RouteID) {
    if (this.loading) {
      this.toastMessage =
        "We're already loading a station's data. Hold on before picking another.";
      throw new Error("Already loading another station's data");
    }

    if (
      this.selectedStop &&
      this.selectedStop.stop.name === stop.name &&
      this.selectedStop.routeID === routeID
    ) {
      this.toastMessage = "You're already at that station.";
      throw new Error("Already at selected station");
    }

    this.selectedStop = { stop, routeID };
    this.headways.length = 0;
    this.dwells.length = 0;
    this.travelTimes.length = 0;
  },

  async selectDestination(stop: Stop, routeID: RouteID) {
    if (this.loading) {
      this.toastMessage =
        "We're already loading a station's data. Hold on before picking another.";
      throw new Error("Already loading another station's data");
    }

    if (!this.selectedStop) {
      throw new Error("No station selected");
    }

    if (
      this.selectedStop.routeID !== routeID &&
      !stop.routeIDs.has(this.selectedStop.routeID)
    ) {
      this.toastMessage =
        "That station isn't on the same route as your selected station.";
      throw new Error("Station not on same route");
    }

    if (
      this.selectedStop.stop.name === stop.name &&
      this.selectedStop.routeID === routeID
    ) {
      this.toastMessage =
        "You can't choose the station you've already selected.";
      throw new Error("Already at selected station");
    }

    this.selectedDestination = {
      stop,
      routeID: this.selectedStop.routeID,
    };
    this.travelTimes.length = 0;
  },

  async fetchData() {
    if (!this.selectedStop) {
      return;
    }

    const stopIDs = reduceStopIDs(this.selectedStop);

    this.loading = true;
    this.loadingMessage = LoadingMessage.None;

    this.error = false;
    this.errorMessage = "";

    try {
      switch (this.selectedCategory) {
        case DataCategory.Headway: {
          if (this.headways.length) {
            break;
          }

          this.loadingMessage = LoadingMessage.Caching;
          await axios.get(`${ENV.VITE_BACKEND_URL}/cache/headway`, {
            params: {
              stop_ids: stopIDs.join(","),
              route_id: this.selectedStop.routeID,
            },
          });

          this.loadingMessage = LoadingMessage.Fetching;
          const res: AxiosResponse = await axios.get(
            `${ENV.VITE_BACKEND_URL}/headway`,
            {
              params: {
                stop_ids: stopIDs.join(","),
                route_id: this.selectedStop.routeID,
              },
            },
          );

          (res.data as BackendResponse<RawHeadway[]>).data.forEach(
            (rawHeadway) => {
              this.headways.push({
                stopID: rawHeadway.stop_id,
                routeID: rawHeadway.route_id,
                prevRouteID: rawHeadway.prev_route_id,
                direction: rawHeadway.direction === "true",
                currentDepDt: dayjs(rawHeadway.current_dep_dt),
                previousDepDt: dayjs(rawHeadway.previous_dep_dt),
                headwayTimeSec: Number(rawHeadway.headway_time_sec),
                benchmarkHeadwayTimeSec: Number(
                  rawHeadway.benchmark_headway_time_sec,
                ),
              });
            },
          );

          this.headways.sort(
            (a, b) => a.currentDepDt.unix() - b.currentDepDt.unix(),
          );
          break;
        }
        case DataCategory.Dwell: {
          if (this.dwells.length) {
            break;
          }

          this.loadingMessage = LoadingMessage.Caching;
          await axios.get(`${ENV.VITE_BACKEND_URL}/cache/dwell`, {
            params: {
              stop_ids: stopIDs.join(","),
              route_id: this.selectedStop.routeID,
            },
          });

          this.loadingMessage = LoadingMessage.Fetching;
          const res: AxiosResponse = await axios.get(
            `${ENV.VITE_BACKEND_URL}/dwell`,
            {
              params: {
                stop_ids: stopIDs.join(","),
                route_id: this.selectedStop.routeID,
              },
            },
          );

          (res.data as BackendResponse<RawDwell[]>).data.forEach((rawDwell) => {
            this.dwells.push({
              stopID: rawDwell.stop_id,
              routeID: rawDwell.route_id,
              direction: rawDwell.direction === "true",
              arrDt: dayjs(rawDwell.arr_dt),
              depDt: dayjs(rawDwell.dep_dt),
              dwellTimeSec: Number(rawDwell.dwell_time_sec),
            });
          });

          this.dwells.sort((a, b) => a.arrDt.unix() - b.arrDt.unix());
          break;
        }
        case DataCategory.TravelTime: {
          if (!this.selectedDestination || this.travelTimes.length) {
            break;
          }

          const toStopIDs = reduceStopIDs(this.selectedDestination);

          this.loadingMessage = LoadingMessage.Caching;
          await axios.get(`${ENV.VITE_BACKEND_URL}/cache/travel_time`, {
            params: {
              from_stop_ids: stopIDs.join(","),
              to_stop_ids: toStopIDs.join(","),
              route_id: this.selectedStop.routeID,
            },
          });

          this.loadingMessage = LoadingMessage.Fetching;
          const res: AxiosResponse = await axios.get(
            `${ENV.VITE_BACKEND_URL}/travel_time`,
            {
              params: {
                from_stop_ids: stopIDs.join(","),
                to_stop_ids: toStopIDs.join(","),
                route_id: this.selectedStop.routeID,
              },
            },
          );

          (res.data as BackendResponse<RawTravelTime[]>).data.forEach(
            (rawTravelTime) => {
              this.travelTimes.push({
                fromStopID: rawTravelTime.from_stop_id,
                toStopID: rawTravelTime.to_stop_id,
                routeID: rawTravelTime.route_id,
                direction: rawTravelTime.direction === "true",
                depDt: dayjs(rawTravelTime.dep_dt),
                arrDt: dayjs(rawTravelTime.arr_dt),
                travelTimeSec: Number(rawTravelTime.travel_time_sec),
                benchmarkTravelTimeSec: Number(
                  rawTravelTime.benchmark_travel_time_sec,
                ),
              });
            },
          );

          this.travelTimes.sort((a, b) => a.depDt.unix() - b.depDt.unix());
          break;
        }
        default:
          throw new Error("Invalid data category");
      }
    } catch (err) {
      this.error = true;
      this.errorMessage = `${err}`;
      throw new Error(`Error fetching data: ${err}`);
    } finally {
      this.loading = false;
      this.loadingMessage = LoadingMessage.None;
    }
  },
});

export default store;
