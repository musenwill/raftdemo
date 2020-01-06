<template>
  <div class="alert container">
    <el-alert
      :title="mes"
      :type="tp"
    />
  </div>
</template>

<script>
export default {
  name: 'Alert',
  props: ['mes', 'tp'],
  data() {
    return {
    }
  }
}
</script>

<style scoped>

</style>
