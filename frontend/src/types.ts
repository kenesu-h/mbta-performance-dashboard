import { Dayjs } from "dayjs";

export interface Route {
  id: RouteID;
  visible: boolean;
  stops: Stop[];
  latLngs: number[][];
}

export enum RouteID {
  Red = "Red",
  Mattapan = "Mattapan",
  Orange = "Orange",
  GreenB = "Green-B",
  GreenC = "Green-C",
  GreenD = "Green-D",
  GreenE = "Green-E",
  Blue = "Blue",
}

export interface Stop {
  ids: string[];
  routeIDs: Set<RouteID>;
  name: string;
  latitude: number;
  longitude: number;
}

export interface RouteStop {
  stop: Stop;
  routeID: RouteID;
}

export enum SelectionMode {
  Normal = "Normal",
  Destination = "Destination",
}

export enum LoadingMessage {
  None = "",
  Caching = "Caching data if we need to, this may take up to a couple seconds...",
  Fetching = "Fetching data...",
}

export enum DataCategory {
  Headway = "Headway",
  Dwell = "Dwell",
  TravelTime = "Travel Time",
}

export interface Headway {
  stopID: string;
  routeID: string;
  prevRouteID: string;
  direction: boolean;
  currentDepDt: Dayjs;
  previousDepDt: Dayjs;
  headwayTimeSec: number;
  benchmarkHeadwayTimeSec: number;
}

export interface Dwell {
  stopID: string;
  routeID: string;
  direction: boolean;
  arrDt: Dayjs;
  depDt: Dayjs;
  dwellTimeSec: number;
}

export interface TravelTime {
  fromStopID: string;
  toStopID: string;
  routeID: string;
  direction: boolean;
  depDt: Dayjs;
  arrDt: Dayjs;
  travelTimeSec: number;
  benchmarkTravelTimeSec: number;
}

export interface BackendResponse<T> {
  data: T;
}

export interface RawShape {
  id: string;
  route_id: string;
  polyline: string;
}

export interface RawStop {
  id: string;
  route_id: string;
  name: string;
  latitude: number;
  longitude: number;
}

export interface RawHeadway {
  stop_id: string;
  route_id: string;
  prev_route_id: string;
  direction: string;
  current_dep_dt: string;
  previous_dep_dt: string;
  headway_time_sec: string;
  benchmark_headway_time_sec: string;
}

export interface RawDwell {
  stop_id: string;
  route_id: string;
  direction: string;
  arr_dt: string;
  dep_dt: string;
  dwell_time_sec: string;
}

export interface RawTravelTime {
  from_stop_id: string;
  to_stop_id: string;
  route_id: string;
  direction: string;
  dep_dt: string;
  arr_dt: string;
  travel_time_sec: string;
  benchmark_travel_time_sec: string;
}
