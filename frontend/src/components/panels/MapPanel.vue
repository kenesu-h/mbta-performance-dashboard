<script setup lang="ts">
import { DropdownChangeEvent } from "primevue/dropdown";
import { useDialog } from "primevue/usedialog";
import { useToast } from "primevue/usetoast";
import { computed } from "vue";

import "leaflet/dist/leaflet.css";
import { LMap, LTileLayer } from "@vue-leaflet/vue-leaflet";

import RouteMarkers from "@/components/RouteMarkers.vue";
import RoutePolyline from "@/components/RoutePolyline.vue";
import MapSettingsDialog from "@/components/dialogs/MapSettingsDialog.vue";
import { MAP_BOUNDS, ROUTE_IDS } from "@/consts";
import appStore from "@/stores/app";
import dataStore from "@/stores/data";
import mapStore from "@/stores/map";
import type { Route, RouteID, Stop } from "@/types";
import { SelectionMode } from "@/types";
import { smallDialogProps } from "@/utils";

const dialog = useDialog();
const toast = useToast();

async function selectStop(stop: Stop, routeID: RouteID) {
  try {
    switch (mapStore.selectionMode) {
      case SelectionMode.Normal:
        await dataStore.selectStop(stop, routeID);
        dataStore.selectedDestination = null;
        break;
      case SelectionMode.Destination:
        await dataStore.selectDestination(stop, routeID);
        appStore.blocked = false;
        mapStore.selectionMode = SelectionMode.Normal;
        break;
      default:
        console.error("Invalid selection mode");
        return;
    }
    try {
      await dataStore.fetchData();
      scrollToDataContainer();
    } catch (err) {
      // Since fetchData sets dataStore.error, let the div in DataPanel render it
    }
  } catch (err) {
    toast.add({
      severity: "error",
      summary: "Error",
      detail: dataStore.toastMessage,
      life: 3000,
    });
  }
}

function scrollToDataContainer() {
  document.getElementById("data-container")?.scrollIntoView({
    behavior: "smooth",
  });
}

const routes = computed(() => {
  return ROUTE_IDS.map((routeID) => {
    return mapStore.routes.get(routeID) as Route;
  });
});

const stops = computed(() => {
  const stops: Stop[] = [];
  mapStore.stops.forEach((stop: Stop) => {
    stops.push(stop);
  });
  return stops;
});
</script>

<template>
  <Panel
    header="Map"
    :pt="{
      root: (_) => ({
        style: {
          width: appStore.width <= 992 ? '100%' : '50%',
          height: '100%',
          display: 'flex',
          'flex-direction': 'column',
          'z-index': 1001,
        },
      }),
      toggleableContent: (_) => ({
        style: {
          'flex-grow': 1,
        },
      }),
      content: (_) => ({
        style: {
          height: '100%',
        },
      }),
    }"
  >
    <div class="flex-column w-100 h-100">
      <l-map
        :use-global-leaflet="false"
        :zoom="11.25"
        :min-zoom="11.25"
        :center="[42.361145, -71.057083]"
        :bounds="MAP_BOUNDS"
        :max-bounds="MAP_BOUNDS"
      >
        <l-tile-layer
          url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
          layer-type="base"
          name="OpenStreetMap"
        />
        <RouteMarkers
          v-for="route in routes"
          :key="route.id"
          :route="route"
          @select-stop="selectStop"
        />
        <RoutePolyline v-for="route in routes" :key="route.id" :route="route" />
      </l-map>
      <Menubar
        :pt="{
          root: (_) => ({
            style: {
              'border-top-left-radius': 0,
              'border-top-right-radius': 0,
            },
          }),
        }"
        :model="[
          {
            label: 'Display Lines',
            icon: 'pi pi-fw pi-cog',
            command: () =>
              dialog.open(MapSettingsDialog, {
                props: smallDialogProps('Display Lines', true),
              }),
          },
        ]"
      >
        <template #end>
          <Dropdown
            filter
            :options="stops"
            optionLabel="name"
            placeholder="Search stops"
            @change="
              async (event: DropdownChangeEvent) => {
                const stop = event.value as Stop;
                if (stop.routeIDs.size) {
                  // Default to the stop's first known route ID
                  selectStop(stop, stop.routeIDs.values().next().value);
                }
              }
            "
          />
        </template>
      </Menubar>
    </div>
  </Panel>
</template>
