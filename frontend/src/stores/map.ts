import { decode } from "@googlemaps/polyline-codec";
import axios, { AxiosResponse } from "axios";
import { reactive } from "vue";

import { ROUTE_IDS } from "@/consts";
import {
  RouteID,
  Route,
  Stop,
  SelectionMode,
  BackendResponse,
  RawShape,
  RawStop,
} from "@/types";
import { strToRouteID } from "@/utils";

const ENV = import.meta.env;

interface Store {
  routes: Map<RouteID, Route>;
  stops: Map<string, Stop>;

  selectionMode: SelectionMode;

  errorMessage: string;

  fetchMapData: () => Promise<void>;
}

const store = reactive<Store>({
  routes: new Map<RouteID, Route>(),
  stops: new Map<string, Stop>(),

  selectionMode: SelectionMode.Normal,

  errorMessage: "",

  async fetchMapData() {
    this.errorMessage = "";

    try {
      {
        const res: AxiosResponse = await axios.get(
          `${ENV.VITE_BACKEND_URL}/shape`,
        );

        (res.data as BackendResponse<RawShape[]>).data.forEach((rawShape) => {
          const route = this.routes.get(strToRouteID(rawShape.route_id));
          if (route) {
            // @ts-ignore
            route.latLngs = [...route.latLngs, decode(rawShape.polyline)];
          }
        });
      }

      {
        const res: AxiosResponse = await axios.get(
          `${ENV.VITE_BACKEND_URL}/stop`,
        );

        (res.data as BackendResponse<RawStop[]>).data.forEach((rawStop) => {
          const routeID: RouteID = strToRouteID(rawStop.route_id);
          const route = this.routes.get(routeID);

          if (route) {
            const stop = this.stops.get(rawStop.name);
            if (stop) {
              stop.ids.push(rawStop.id);
              stop.routeIDs.add(routeID);
            } else {
              const stop = {
                ids: [rawStop.id],
                routeIDs: new Set<RouteID>([routeID]),
                name: rawStop.name,
                latitude: rawStop.latitude,
                longitude: rawStop.longitude,
              };
              route.stops.push(stop);
              this.stops.set(rawStop.name, stop);
            }
          }
        });
      }
    } catch (err) {
      this.errorMessage = `${err}`;
      throw new Error(`Error fetching map data: ${err}`);
    }
  },
});

ROUTE_IDS.forEach((routeID) => {
  store.routes.set(routeID, {
    id: routeID,
    visible: true,
    stops: [],
    latLngs: [],
  });
});

export default store;
