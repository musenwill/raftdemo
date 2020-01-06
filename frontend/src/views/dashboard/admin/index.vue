<template>
<!--
  <div class="dashboard-editor-container">
    <github-corner class="github-corner" />
    <panel-group v-bind:clusterList='clusterList'/>
  </div>
-->
</template>

<script>
import GithubCorner from '@/components/GithubCorner'
import PanelGroup from './components/PanelGroup'
import echarts from 'echarts'

export default {
  name: 'DashboardAdmin',
  components: {
    GithubCorner,
    PanelGroup
  },
  created: function() {
    console.log('created')
  },
  data() {
    console.log('data ok')
    return {
      clusterList: null,
    }
  },
  methods: {
  }
}
</script>

<style lang="scss" scoped>
.dashboard-editor-container {
  padding: 32px;
  background-color: rgb(240, 242, 245);
  position: relative;

  .github-corner {
    position: absolute;
    top: 0px;
    border: 0;
    right: 0;
  }

  .chart-wrapper {
    background: #fff;
    padding: 16px 16px 0;
    margin-bottom: 32px;
  }
}
</style>
