<script setup lang="ts">
import { ref } from "vue";

const terms = ref([
  {
    term: "Open",
    definition:
      "In a candlestick graph, this is the value of the data point at the start of the " +
      "period.",
  },
  {
    term: "High",
    definition:
      "In a candlestick graph, this is the highest value of all data points during the " +
      "whole period.",
  },
  {
    term: "Low",
    definition:
      "In a candlestick graph, this is the lowest value of all data points during the " +
      "whole period.",
  },
  {
    term: "Close",
    definition:
      "In a candlestick graph, this is the value of the data point at the end of the " +
      "period.",
  },
  {
    term: "Headway",
    definition:
      "The time between the departures of the previous and current trains at a stop.",
  },
  {
    term: "Benchmark Headway",
    definition:
      "The average scheduled time between the departures of the previous and current " +
      "trains at a stop.",
  },
  {
    term: "Dwell",
    definition: "The time a train spent stationary at a stop.",
  },
]);
</script>

<template>
  <DataTable :value="terms">
    <Column field="term" header="Term" />
    <Column field="definition" header="Definition" />
  </DataTable>
</template>
