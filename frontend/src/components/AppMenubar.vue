<script setup lang="ts">
import { useDialog } from "primevue/usedialog";

import AboutDialog from "@/components/dialogs/AboutDialog.vue";
import UpdatesDialog from "@/components/dialogs/UpdatesDialog.vue";
import { largeDialogProps } from "@/utils";

const dialog = useDialog();
</script>

<template>
  <Menubar
    :pt="{
      root: (_) => ({
        style: {
          'z-index': 1002,
        },
      }),
    }"
    :model="[
      {
        label: 'About',
        icon: 'pi pi-fw pi-info-circle',
        command: () =>
          dialog.open(AboutDialog, { props: largeDialogProps('About', true) }),
      },
      {
        label: 'Updates',
        icon: 'pi pi-fw pi-history',
        command: () =>
          dialog.open(UpdatesDialog, {
            props: largeDialogProps('Updates', true),
          }),
      },
    ]"
  >
    <template #start>
      <strong :style="{ margin: '0 1rem 0 1rem' }"
        >MBTA Performance Dashboard</strong
      >
    </template>
    <template #item="{ label, props }">
      <a v-bind="props.action">
        <span v-bind="props.icon" />
        <span v-bind="props.label">{{ label }}</span>
      </a>
    </template>
    <template #end>
      <a href="https://github.com/kenesu-h/mbta-performance-dashboard">
        <span
          class="pi pi-github"
          :style="{
            color: 'white',
            'font-size': '1.5rem',
            'margin-right': '1rem',
          }"
        />
      </a>
    </template>
  </Menubar>
</template>
