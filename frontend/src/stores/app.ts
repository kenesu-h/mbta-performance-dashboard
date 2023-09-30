import { reactive } from "vue";

interface Store {
  blocked: boolean;
  width: number;
}

const store = reactive<Store>({
  blocked: false,
  width: 0,
});

export default store;
