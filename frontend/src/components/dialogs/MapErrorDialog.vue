<script setup lang="ts">
import { DynamicDialogInstance } from "primevue/dynamicdialogoptions";
import { useDialog } from "primevue/usedialog";
import { inject, Ref } from "vue";

import MapErrorDialog from "@/components/dialogs/MapErrorDialog.vue";
import MapLoadingDialog from "@/components/dialogs/MapLoadingDialog.vue";
import store from "@/stores/map";
import { mediumDialogProps, smallDialogProps } from "@/utils";

const dialog = useDialog();
const dialogRef = inject("dialogRef") as Ref<DynamicDialogInstance>;
function closeDialog() {
  dialogRef.value.close();
}

async function handleClick() {
  closeDialog();
  const loadingDialogRef = dialog.open(MapLoadingDialog, {
    props: smallDialogProps("Loading", false),
  });

  try {
    await store.fetchMapData();
    loadingDialogRef.close();
  } catch (err) {
    loadingDialogRef.close();
    dialog.open(MapErrorDialog, { props: mediumDialogProps("Error", false) });
  }
}
</script>

<template>
  <p>An error occurred while fetching map data:</p>
  <Message severity="error" :closable="false">{{ store.errorMessage }}</Message>
  <div class="flex-row" :style="{ 'justify-content': 'end' }">
    <Button label="Retry" @click="handleClick" />
  </div>
</template>
