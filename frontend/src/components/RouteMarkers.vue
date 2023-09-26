<script setup lang="ts">
import { LMarker, LTooltip } from "@vue-leaflet/vue-leaflet";
import { ref, computed } from "vue";

import { Route } from "@/types";

const props = defineProps<{
  route: Route;
}>();
const emit = defineEmits(["selectStop"]);

const route = ref(props.route);
const stops = computed(() => {
  return route.value.visible ? route.value.stops || [] : [];
});
</script>

<template>
  <l-marker
    v-for="stop in stops"
    :key="stop.ids[0]"
    :lat-lng="[stop.latitude, stop.longitude]"
    @click="emit('selectStop', stop, route.id)"
  >
    <l-tooltip>{{ stop.name }}</l-tooltip>
  </l-marker>
</template>
