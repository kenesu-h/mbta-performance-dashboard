import { createApp } from "vue";
import "@/style.css";
import App from "@/App.vue";

import "primevue/resources/themes/soho-dark/theme.css";
import "primeicons/primeicons.css";

import PrimeVue from "primevue/config";

import Badge from "primevue/badge";
import BlockUI from "primevue/blockui";
import Button from "primevue/button";
import Checkbox from "primevue/checkbox";
import Column from "primevue/column";
import DataTable from "primevue/datatable";
import Dialog from "primevue/dialog";
import DialogService from "primevue/dialogservice";
import DynamicDialog from "primevue/dynamicdialog";
import InputNumber from "primevue/inputnumber";
import InputText from "primevue/inputtext";
import Menubar from "primevue/menubar";
import Message from "primevue/message";
import Panel from "primevue/panel";
import ProgressSpinner from "primevue/progressspinner";
import Slider from "primevue/slider";
import TabView from "primevue/tabview";
import TabPanel from "primevue/tabpanel";
import Toast from "primevue/toast";
import ToastService from "primevue/toastservice";
import ToggleButton from "primevue/togglebutton";
import Tooltip from "primevue/tooltip";
import VueApexCharts from "vue3-apexcharts";

const app = createApp(App);

app.use(PrimeVue);
app.use(DialogService);
app.use(ToastService);

app.component("Badge", Badge);
app.component("BlockUI", BlockUI);
app.component("Button", Button);
app.component("Checkbox", Checkbox);
app.component("Column", Column);
app.component("DataTable", DataTable);
app.component("Dialog", Dialog);
app.component("DynamicDialog", DynamicDialog);
app.component("InputNumber", InputNumber);
app.component("InputText", InputText);
app.component("Menubar", Menubar);
app.component("Message", Message);
app.component("Panel", Panel);
app.component("ProgressSpinner", ProgressSpinner);
app.component("Slider", Slider);
app.component("TabView", TabView);
app.component("TabPanel", TabPanel);
app.component("Toast", Toast);
app.component("ToggleButton", ToggleButton);
app.component("ApexChart", VueApexCharts);

app.directive("tooltip", Tooltip);

app.mount("#app");
