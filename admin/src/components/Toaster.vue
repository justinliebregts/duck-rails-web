<template>
  <CToaster position="top-right">
    <CToast
      v-for="toast in toasts"
      :key="toast.id"
      :show="true"
      :color="toast.color"
      :autohide="2000"
    >
      {{ toast.msg }}
    </CToast>
  </CToaster>
</template>

<script>
export default {
  name: "Toaster",
  props: {
    toasts: {
      type: Array,
      default: () => []
    }
  },
  methods:{       
    toast(msg, color = "primary") {
        // Generate a unique ID (timestamp + random number)
        const newToast = {
            id: Date.now() + Math.random(),
            msg: msg,
            color: color
        };
  
        // Reactive array push
        this.toasts.push(newToast);
    } 
  }
};
</script>