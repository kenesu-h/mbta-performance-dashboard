<script setup lang="ts">
import { LPolyline } from "@vue-leaflet/vue-leaflet";
import { ref, computed } from "vue";

import type { Route } from "@/types";
import { routeIDToColor } from "@/utils";

const props = defineProps<{
  route: Route;
}>();

const route = ref(props.route);
const color = computed(() => {
  return routeIDToColor(route.value.id);
});
const latLngs = computed(() => {
  return route.value.visible ? route.value.latLngs || [] : [];
});
</script>

<template>
  <l-polyline :color="color" :lat-lngs="latLngs" />
</template>
