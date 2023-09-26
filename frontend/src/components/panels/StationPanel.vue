<script setup lang="ts">
import { useToast } from "primevue/usetoast";
import { computed } from "vue";

import { PANEL_STYLE } from "@/consts";
import dataStore from "@/stores/data";
import mapStore from "@/stores/map";
import { RouteID, RouteStop, SelectionMode } from "@/types";

const toast = useToast();

const routeIDs = computed(() => {
  return dataStore.selectedStop ? dataStore.selectedStop.stop.routeIDs : [];
});

function isRouteSelected(routeID: RouteID): boolean {
  return (
    Boolean(dataStore.selectedStop) &&
    (dataStore.selectedStop as RouteStop).routeID === routeID
  );
}
</script>

<template>
  <Panel
    :header="`${dataStore.selectedStop?.stop.name || ''} Station`"
    :pt="PANEL_STYLE"
  >
    <div class="flex-row gap">
      <Button
        v-for="routeID in routeIDs"
        :key="routeID"
        :label="routeID"
        :disabled="dataStore.selectedStop?.routeID === routeID"
        @click="
          async () => {
            if (dataStore.selectedStop && !isRouteSelected(routeID)) {
              try {
                await dataStore.selectStop(
                  dataStore.selectedStop.stop,
                  routeID,
                );
                dataStore.selectedDestination = null;
                mapStore.selectionMode = SelectionMode.Normal;
                try {
                  await dataStore.fetchData();
                } catch (err) {
                  // Since fetchData sets dataStore.error, let the div in DataPanel render it
                }
              } catch (err) {
                toast.add({
                  severity: 'error',
                  summary: 'Error',
                  detail: dataStore.toastMessage,
                  life: 3000,
                });
              }
            }
          }
        "
      />
    </div>
  </Panel>
</template>
