<script setup lang="ts">
import { useDialog } from "primevue/usedialog";
import { computed, onMounted } from "vue";

import "leaflet/dist/leaflet.css";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";

import AppMenubar from "@/components/AppMenubar.vue";
import MapErrorDialog from "@/components/dialogs/MapErrorDialog.vue";
import MapLoadingDialog from "@/components/dialogs/MapLoadingDialog.vue";
import MobileWarningDialog from "@/components/dialogs/MobileWarningDialog.vue";
import DataPanel from "@/components/panels/DataPanel.vue";
import MapPanel from "@/components/panels/MapPanel.vue";
import StationPanel from "@/components/panels/StationPanel.vue";
import appStore from "@/stores/app";
import dataStore from "@/stores/data";
import mapStore from "@/stores/map";
import { mediumDialogProps, smallDialogProps } from "@/utils";

dayjs.extend(utc);
dayjs.extend(timezone);

const dialog = useDialog();

onMounted(async () => {
  const loadingDialogRef = dialog.open(MapLoadingDialog, {
    props: smallDialogProps("Loading", false),
  });

  try {
    await mapStore.fetchMapData();
    loadingDialogRef.close();

    if (window.innerWidth <= 768) {
      dialog.open(MobileWarningDialog, {
        props: mediumDialogProps("Warning", true),
      });
    }
  } catch (err) {
    loadingDialogRef.close();
    dialog.open(MapErrorDialog, { props: mediumDialogProps("Error", false) });
  }
});

const isMobile = computed(() => {
  return window.innerWidth <= 768;
});
</script>

<template>
  <Toast />
  <DynamicDialog />

  <!-- Block UI for when selecting a destination for travel times -->
  <BlockUI
    :blocked="appStore.blocked"
    :base-z-index="1000"
    :auto-z-index="false"
    full-screen
  />

  <AppMenubar />

  <div class="flex-row gap h-100" :class="{ 'flex-wrap': isMobile }">
    <MapPanel />

    <div
      v-if="dataStore.selectedStop"
      class="flex-column flex-grow gap h-100"
      :style="{ 'justify-content': 'start' }"
    >
      <StationPanel />
      <DataPanel />
    </div>
    <div
      v-else
      class="flex-column flex-grow justify-center text-align-center h-100"
    >
      Click on a station on the map to get started!
    </div>
  </div>
</template>

<style>
.leaflet-container {
  border-radius: 6px 6px 0 0;
}

.w-100 {
  width: 100%;
}

.h-100 {
  height: 100%;
}

.text-align-center {
  text-align: center;
}

.flex-row {
  display: flex;
  flex-direction: row;
}

.flex-column {
  display: flex;
  flex-direction: column;
}

.flex-grow {
  flex-grow: 1;
}

.flex-wrap {
  flex-wrap: wrap;
}

.justify-center {
  justify-content: center;
}

.gap {
  gap: 0.5rem;
}
</style>
