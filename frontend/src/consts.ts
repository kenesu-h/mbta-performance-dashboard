// @ts-ignore
// Reason: https://github.com/vue-leaflet/vue-leaflet/issues/48#issuecomment-733774307
import { latLngBounds } from "leaflet/dist/leaflet-src.esm";

import { RouteID } from "@/types";

export const ENV = import.meta.env;

export const MAP_BOUNDS = latLngBounds(
  [42.192049, -71.264843],
  [42.476643, -70.949545],
);

export const DETAILED_VIEW = "Toggle Detailed View";

export const PANEL_STYLE = {
  root: (_: any) => ({
    style: {
      width: "100%",
    },
  }),
};

export const ROUTE_IDS = [
  RouteID.Red,
  RouteID.Mattapan,
  RouteID.Orange,
  RouteID.GreenB,
  RouteID.GreenC,
  RouteID.GreenD,
  RouteID.GreenE,
  RouteID.Blue,
];
