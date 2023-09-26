<script setup lang="ts">
import { ROUTE_IDS } from "@/consts";
import store from "@/stores/map";
import type { Route } from "@/types";
</script>

<template>
  <div class="flex-column gap">
    <div v-for="routeID in ROUTE_IDS" :key="routeID">
      <Checkbox
        v-model="(store.routes.get(routeID) as Route).visible"
        :input-id="routeID"
        :name="routeID"
        :binary="true"
      />
      <label :for="routeID" :style="{ 'margin-left': '1rem' }">{{
        routeID
      }}</label>
    </div>
  </div>
</template>
