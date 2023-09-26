import { DialogProps } from "primevue/dialog";

import { RouteID, RouteStop } from "@/types";

export function strToRouteID(str: string): RouteID {
  switch (str) {
    case "Red":
      return RouteID.Red;
    case "Mattapan":
      return RouteID.Mattapan;
    case "Orange":
      return RouteID.Orange;
    case "Green-B":
      return RouteID.GreenB;
    case "Green-C":
      return RouteID.GreenC;
    case "Green-D":
      return RouteID.GreenD;
    case "Green-E":
      return RouteID.GreenE;
    case "Blue":
      return RouteID.Blue;
    default:
      throw new Error("Invalid route ID");
  }
}

export function routeIDToColor(routeID: RouteID): string {
  switch (routeID) {
    case RouteID.Red:
      return "#d20f39";
    case RouteID.Mattapan:
      return "#d20f39";
    case RouteID.Orange:
      return "#fe640b";
    case RouteID.GreenB:
      return "#40a02b";
    case RouteID.GreenC:
      return "#40a02b";
    case RouteID.GreenD:
      return "#40a02b";
    case RouteID.GreenE:
      return "#40a02b";
    case RouteID.Blue:
      return "#1e66f5";
    default:
      throw new Error("Invalid route ID");
  }
}

export function reduceStopIDs(routeStop: RouteStop): string[] {
  return routeStop.stop.ids.reduce(
    (acc: string[], id: string) => [...acc, id],
    [],
  );
}

function baseDialogProps(header: string, closable: boolean): DialogProps {
  return {
    header,
    breakpoints: {
      "768px": "90vw",
    },
    modal: true,
    closable,
  };
}

export function smallDialogProps(
  header: string,
  closable: boolean,
): DialogProps {
  return {
    ...baseDialogProps(header, closable),
    style: {
      width: "16vw",
    },
  };
}

export function mediumDialogProps(
  header: string,
  closable: boolean,
): DialogProps {
  return {
    ...baseDialogProps(header, closable),
    style: {
      width: "33vw",
    },
  };
}

export function largeDialogProps(
  header: string,
  closable: boolean,
): DialogProps {
  return {
    ...baseDialogProps(header, closable),
    style: {
      width: "50vw",
    },
  };
}
