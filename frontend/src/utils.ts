import dayjs, { Dayjs } from "dayjs";
import timezone from "dayjs/plugin/timezone";
import { DialogProps } from "primevue/dialog";

import { RouteID, RouteStop } from "@/types";

dayjs.extend(timezone);

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

function calculateChunks<T>(
  data: T[],
  x: (t: T) => Dayjs,
  period: number,
): T[][] {
  if (!data.length) {
    return [];
  }

  const chunks: T[][] = [];
  let i = 0;
  let startOfRange = dayjs()
    .tz("America/New_York")
    .startOf("day")
    .subtract(30, "days");
  let endOfRange = startOfRange.add(period, "hours");

  data.forEach((t) => {
    const date = dayjs(x(t));

    if (i >= chunks.length) {
      chunks.push([]);
    }

    if (
      (date.isSame(startOfRange) || date.isAfter(startOfRange)) &&
      date.isBefore(endOfRange)
    ) {
      chunks[i].push(t);
    } else {
      startOfRange = startOfRange.add(period, "hours");
      endOfRange = endOfRange.add(period, "hours");
      i++;
    }
  });

  return chunks;
}

export function calculateCandlesticks<T>(
  data: T[],
  x: (t: T) => Dayjs,
  y: (t: T) => number,
  period: number,
): { x: number; y: number[] }[] {
  if (!data.length) {
    return [];
  }

  const candlesticks: { x: number; y: number[] }[] = [];
  const chunks: T[][] = calculateChunks(data, x, period);

  chunks.forEach((chunk) => {
    if (chunk.length) {
      const open = y(chunk[0]);
      let high = y(chunk[0]);
      let low = y(chunk[0]);
      const close = y(chunk[chunk.length - 1]);

      chunk.forEach((t) => {
        high = high ? Math.max(high, y(t)) : y(t);
        low = low ? Math.min(low, y(t)) : y(t);
      });

      const timestamp = x(chunk[0]).unix();
      candlesticks.push({
        x: timestamp,
        y: [
          parseFloat(open.toFixed(2)),
          parseFloat(high.toFixed(2)),
          parseFloat(low.toFixed(2)),
          parseFloat(close.toFixed(2)),
        ],
      });
    }
  });

  return candlesticks;
}

export function calculateAverages<T>(
  data: T[],
  x: (t: T) => Dayjs,
  y: (t: T) => number,
  period: number,
): { x: number; y: number }[] {
  if (!data.length) {
    return [];
  }

  const averages: { x: number; y: number }[] = [];
  const chunks: T[][] = calculateChunks(data, x, period);

  chunks.forEach((chunk) => {
    if (chunk.length) {
      const sum = chunk.reduce((acc, t) => {
        return acc + y(t);
      }, 0);

      const timestamp = x(chunk[0]).unix();
      averages.push({
        x: timestamp,
        y: parseFloat((sum / chunk.length).toFixed(2)),
      });
    }
  });

  return averages;
}
