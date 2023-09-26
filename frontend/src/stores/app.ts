import { reactive } from "vue";

interface Store {
  blocked: boolean;
}

const store = reactive<Store>({
  blocked: false,
});

export default store;
