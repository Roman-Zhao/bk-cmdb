<template>
  <section class="tree-layout" v-bkloading="{ isLoading: $loading(Object.values(request)) }">
    <bk-input class="tree-search" v-test-id
      clearable
      right-icon="bk-icon icon-search"
      :placeholder="$t('请输入关键词')"
      v-model.trim="filter">
    </bk-input>
    <bk-big-tree ref="tree" class="topology-tree" v-test-id
      selectable
      display-matched-node-descendants
      :height="$APP.height - 160"
      :node-height="36"
      :options="{
        idKey: getNodeId,
        nameKey: 'bk_inst_name',
        childrenKey: 'child'
      }"
      @select-change="handleSelectChange"
      @expand-change="handleExpandChange">
      <div :class="['node-info clearfix', { 'is-selected': node.selected }]" slot-scope="{ node, data }">
        <i class="internal-node-icon fl"
          v-if="data.default !== 0"
          :class="getInternalNodeClass(node, data)">
        </i>
        <i v-else
          :class="['node-icon fl', { 'is-selected': node.selected, 'is-template': isTemplate(node) }]">
          {{data.bk_obj_name[0]}}
        </i>
        <cmdb-auth v-if="showCreate(node, data)"
          class="info-create-trigger fr"
          :auth="{ type: $OPERATION.C_TOPO, relation: [bizId] }">
          <template slot-scope="{ disabled }">
            <i v-if="isBlueKing && !editable"
              class="node-button disabled-node-button"
              v-bk-tooltips.top="{ content: $t('蓝鲸业务拓扑节点提示'), interactive: false }">
              {{$t('新建')}}
            </i>
            <i v-else-if="data.set_template_id"
              class="node-button disabled-node-button"
              v-bk-tooltips.top="{
                content: getSetNodeTips(node),
                interactive: true,
                onShow: handleSetNodeTipsToggle,
                onHide: handleSetNodeTipsToggle
              }">
              {{$t('新建')}}
            </i>
            <bk-button v-else class="node-button" v-test-id="'createNode'"
              theme="primary"
              :disabled="disabled"
              @click.stop="showCreateDialog(node)">
              {{$t('新建')}}
            </bk-button>
          </template>
        </cmdb-auth>
        <cmdb-loading :class="['node-count fr', { 'is-selected': node.selected }]"
          :loading="['pending', undefined].includes(data.status)">
          {{getNodeCount(data)}}
        </cmdb-loading>
        <span class="node-name" :title="node.name">{{node.name}}</span>
      </div>
    </bk-big-tree>
    <bk-dialog class="bk-dialog-no-padding"
      v-model="createInfo.show"
      :show-footer="false"
      :mask-close="false"
      :width="580"
      @after-leave="handleAfterCancelCreateNode"
      @cancel="handleCancelCreateNode">
      <template v-if="createInfo.nextModelId === 'module'">
        <create-module v-if="createInfo.visible"
          :parent-node="createInfo.parentNode"
          @submit="handleCreateNode"
          @cancel="handleCancelCreateNode">
        </create-module>
      </template>
      <template v-else-if="createInfo.nextModelId === 'set'">
        <create-set v-if="createInfo.visible"
          :parent-node="createInfo.parentNode"
          @submit="handleCreateSetNode"
          @cancel="handleCancelCreateNode">
        </create-set>
      </template>
      <template v-else>
        <create-node v-if="createInfo.visible"
          :next-model-id="createInfo.nextModelId"
          :properties="createInfo.properties"
          :parent-node="createInfo.parentNode"
          @submit="handleCreateNode"
          @cancel="handleCancelCreateNode">
        </create-node>
      </template>
    </bk-dialog>
  </section>
</template>

<script>
  import { mapGetters } from 'vuex'
  import debounce from 'lodash.debounce'
  import CreateNode from './create-node.vue'
  import CreateSet from './create-set.vue'
  import CreateModule from './create-module.vue'
  import Bus from '@/utils/bus'
  import RouterQuery from '@/router/query'
  import { addResizeListener, removeResizeListener } from '@/utils/resize-events'
  import FilterStore from '@/components/filters/store'
  import CmdbLoading from '@/components/loading/loading'
  import { sortTopoTree } from '@/utils/tools'
  import {
    MENU_BUSINESS_HOST_AND_SERVICE
  } from '@/dictionary/menu-symbol'
  export default {
    components: {
      CreateNode,
      CreateSet,
      CreateModule,
      CmdbLoading
    },
    props: {
      active: {
        type: String,
        required: true
      }
    },
    data() {
      return {
        isBlueKing: false,
        filter: RouterQuery.get('keyword', ''),
        handleFilter: () => ({}),
        nodeCountType: 'host_count',
        nodeIconMap: {
          1: 'icon-cc-host-free-pool',
          2: 'icon-cc-host-breakdown',
          default: 'icon-cc-host-free-pool'
        },
        request: {
          instance: Symbol('instance'),
          internal: Symbol('internal'),
          property: Symbol('property')
        },
        createInfo: {
          show: false,
          visible: false,
          properties: [],
          parentNode: null,
          nextModelId: null
        },
        editable: false,
        timer: null
      }
    },
    computed: {
      ...mapGetters('objectBiz', ['bizId']),
      ...mapGetters('businessHost', ['topologyModels', 'propertyMap']),
      ...mapGetters('businessHost', ['selectedNode'])
    },
    watch: {
      filter(value) {
        this.handleFilter()
        RouterQuery.set('keyword', value)
      },
      active: {
        immediate: true,
        handler(value) {
          const map = {
            hostList: 'host_count',
            serviceInstance: 'service_instance_count'
          }
          if (Object.keys(map).includes(value)) {
            this.nodeCountType = map[value]
          }
        }
      },
      isBlueKing(flag) {
        if (flag) {
          this.getBlueKingEditStatus()
          clearInterval(this.timer)
          this.timer = setInterval(this.getBlueKingEditStatus, 1000 * 60)
        }
      }
    },
    created() {
      Bus.$on('refresh-count', this.refreshCount)
      Bus.$on('refresh-count-by-node', this.refreshCountByNode)
      this.initTopology()
    },
    mounted() {
      addResizeListener(this.$el, this.handleResize)
    },
    beforeDestroy() {
      this.destroyWatcher()
      Bus.$off('refresh-count', this.refreshCount)
      Bus.$off('refresh-count-by-node', this.refreshCountByNode)
      clearInterval(this.timer)
      removeResizeListener(this.$el, this.handleResize)
    },
    methods: {
      async initTopology() {
        try {
          const [topology, internal] = await Promise.all([
            this.getInstanceTopology(),
            this.getInternalTopology()
          ])
          sortTopoTree(topology, 'bk_inst_name', 'child')
          sortTopoTree(internal.module, 'bk_module_name')
          const root = topology[0] || {}
          const children = root.child || []
          const idlePool = {
            bk_obj_id: 'set',
            bk_inst_id: internal.bk_set_id,
            bk_inst_name: internal.bk_set_name,
            default: internal.default,
            is_idle_set: true,
            child: internal.module.map(module => ({
              bk_obj_id: 'module',
              bk_inst_id: module.bk_module_id,
              bk_inst_name: module.bk_module_name,
              default: module.default
            }))
          }
          children.unshift(idlePool)
          this.isBlueKing = root.bk_inst_name === '蓝鲸'
          this.$refs.tree.setData(topology)
          this.createWatcher()
        } catch (e) {
          console.error(e)
        }
      },
      createWatcher() {
        this.nodeUnwatch = RouterQuery.watch('node', this.setDefaultState, { immediate: true })
        this.filterUnwatch = RouterQuery.watch('keyword', (value) => {
          this.filter = value
        })
        this.handleFilter = debounce(() => {
          this.$refs.tree.filter(this.filter)
          this.filter && this.setNodeCount(this.$refs.tree.visibleNodes)
        }, 300)
      },
      destroyWatcher() {
        this.nodeUnwatch && this.nodeUnwatch()
        this.filterUnwatch && this.filterUnwatch()
      },
      setDefaultState() {
        // 非业务拓扑主页面不触发设置节点选中等，防止查询条件非预期的被清除
        if (this.$route.name !== MENU_BUSINESS_HOST_AND_SERVICE) {
          return
        }
        const defaultNode = this.getDefaultNode()
        if (defaultNode) {
          const { tree } = this.$refs
          tree.setExpanded(defaultNode.id)
          tree.setSelected(defaultNode.id, { emitEvent: true })
          this.handleDefaultExpand(defaultNode)
          // 仅对第一次设置时调整滚动位置
          !this.initialized && this.$nextTick(() => {
            this.initialized = true
            const index = tree.visibleNodes.indexOf(defaultNode)
            tree.$refs.virtualScroll.scrollPageByIndex(index)
          })
        }
      },
      getDefaultNode() {
        // 选中指定的节点
        const queryNodeId = RouterQuery.get('node', '')
        if (queryNodeId) {
          const node = this.$refs.tree.getNodeById(queryNodeId)
          if (node) {
            return node
          }
        }
        // 从其他页面跳转过来需要筛选节点，例如：删除集群模板中的服务模板
        const keyword = RouterQuery.get('keyword', '')
        if (keyword) {
          const [firstMatchedNode] = this.$refs.tree.filter(keyword.trim())
          if (firstMatchedNode) {
            return firstMatchedNode
          }
        }
        // 选中第一个节点
        const [firstNode] = this.$refs.tree.nodes
        return firstNode || null
      },
      getInstanceTopology() {
        return this.$store.dispatch('objectMainLineModule/getInstTopoInstanceNum', {
          bizId: this.bizId,
          config: {
            requestId: this.request.instance
          }
        })
      },
      getInternalTopology() {
        return this.$store.dispatch('objectMainLineModule/getInternalTopo', {
          bizId: this.bizId,
          config: {
            requestId: this.request.internal
          }
        })
      },
      getNodeId(data) {
        return `${data.bk_obj_id}-${data.bk_inst_id}`
      },
      getInternalNodeClass(node, data) {
        const clazz = []
        clazz.push(this.nodeIconMap[data.default] || this.nodeIconMap.default)
        if (node.selected) {
          clazz.push('is-selected')
        }
        return clazz
      },
      handleSelectChange(node) {
        this.$store.commit('businessHost/setSelectedNode', node)
        Bus.$emit('toggle-host-filter', false)
        const query = {
          node: node.id,
          page: 1,
          _t: Date.now()
        }
        RouterQuery.set(query)
        this.initialized && FilterStore.setActiveCollection(null)
      },
      handleDefaultExpand(node) {
        const nodes = []
        let parentNode = node
        while (parentNode) {
          nodes.push(...parentNode.children)
          if (!parentNode.parent) {
            nodes.push(parentNode)
          }
          parentNode = parentNode.parent
        }
        this.setNodeCount(nodes)
      },
      handleExpandChange(node) {
        if (!node.expanded) return
        this.setNodeCount([node, ...node.children])
      },
      async setNodeCount(targetNodes, force = false) {
        const nodes = force
          ? targetNodes
          : targetNodes.filter(({ data }) => !['pending', 'finished'].includes(data.status))
        if (!nodes.length) return
        nodes.forEach(({ data }) => this.$set(data, 'status', 'pending'))
        try {
          const result = await this.$store.dispatch('objectMainLineModule/getTopoStatistics', {
            bizId: this.bizId,
            params: {
              condition: nodes.map(({ data }) => ({ bk_obj_id: data.bk_obj_id, bk_inst_id: data.bk_inst_id }))
            }
          })
          nodes.forEach(({ data }) => {
            // eslint-disable-next-line
            const count = result.find(count => count.bk_obj_id === data.bk_obj_id && count.bk_inst_id === data.bk_inst_id)
            this.$set(data, 'status', 'finished')
            this.$set(data, 'host_count', count.host_count)
            this.$set(data, 'service_instance_count', count.service_instance_count)
          })
        } catch (error) {
          console.error(error)
          nodes.forEach((node) => {
            this.$set(node.data, 'status', 'error')
          })
        }
      },
      getNodeCount(data) {
        const count = data[this.nodeCountType]
        if (typeof count === 'number') {
          return count
        }
        return 0
      },
      showCreate(node, data) {
        const isModule = data.bk_obj_id === 'module'
        const isIdleSet = data.is_idle_set
        return !isModule && !isIdleSet
      },
      async getBlueKingEditStatus() {
        try {
          this.editable = await this.$store.dispatch('getBlueKingEditStatus', {
            config: {
              globalError: false
            }
          })
          this.$store.commit('businessHost/setBlueKingEditable', this.editable)
        } catch (_) {
          this.editable = false
        }
      },
      getSetNodeTips(node) {
        const tips = document.createElement('div')
        const span = document.createElement('span')
        span.innerText = this.$t('需在集群模板中新建')
        const link = document.createElement('a')
        link.innerText = this.$t('立即跳转')
        link.href = 'javascript:void(0)'
        link.style.color = '#3a84ff'
        link.addEventListener('click', () => {
          this.$routerActions.redirect({
            name: 'setTemplateConfig',
            params: {
              mode: 'view',
              templateId: node.data.set_template_id
            },
            history: true
          })
        })
        tips.appendChild(span)
        tips.appendChild(link)
        return tips
      },
      handleSetNodeTipsToggle(tips) {
        const element = tips.reference.parentElement
        if (tips.state.isVisible) {
          element.classList.remove('hovering')
        } else {
          element.classList.add('hovering')
        }
        return true
      },
      async showCreateDialog(node) {
        const nodeModel = this.topologyModels.find(data => data.bk_obj_id === node.data.bk_obj_id)
        const nextModelId = nodeModel.bk_next_obj
        this.createInfo.nextModelId = nextModelId
        this.createInfo.parentNode = node
        this.createInfo.show = true
        this.createInfo.visible = true
        let properties = this.propertyMap[nextModelId]
        if (!properties) {
          const action = 'objectModelProperty/searchObjectAttribute'
          properties = await this.$store.dispatch(action, {
            params: {
              bk_biz_id: this.bizId,
              bk_obj_id: nextModelId,
              bk_supplier_account: this.$store.getters.supplierAccount
            },
            config: {
              requestId: this.request.property
            }
          })
          if (!['set', 'module'].includes(nextModelId)) {
            this.$store.commit('businessHost/setProperties', {
              id: nextModelId,
              properties
            })
          }
        }
        const primaryKey = { set: 'bk_set_id', module: 'bk_module_id' }[nextModelId] || 'bk_inst_id'
        this.createInfo.properties = properties.filter(property => property.bk_property_id !== primaryKey)
      },
      handleAfterCancelCreateNode() {
        this.createInfo.visible = false
        this.createInfo.properties = []
        this.createInfo.parentNode = null
        this.createInfo.nextModelId = null
      },
      handleCancelCreateNode() {
        this.createInfo.show = false
      },
      async handleCreateNode(value) {
        try {
          const { parentNode } = this.createInfo
          const formData = {
            ...value,
            bk_biz_id: this.bizId,
            bk_parent_id: parentNode.data.bk_inst_id
          }
          const { nextModelId } = this.createInfo
          const nextModel = this.topologyModels.find(model => model.bk_obj_id === nextModelId)
          const handlerMap = {
            set: this.createSet,
            module: this.createModule
          }
          const data = await (handlerMap[nextModelId] || this.createCommonInstance)(formData)
          const nodeData = {
            default: 0,
            child: [],
            bk_obj_name: nextModel.bk_obj_name,
            bk_obj_id: nextModel.bk_obj_id,
            host_count: 0,
            service_instance_count: 0,
            service_template_id: value.service_template_id,
            status: 'finished',
            ...data
          }
          this.$refs.tree.addNode(nodeData, parentNode.id, parentNode.data.bk_obj_id === 'biz' ? 1 : 0)
          this.$success(this.$t('新建成功'))
          this.createInfo.show = false
        } catch (e) {
          console.error(e)
        }
      },
      async handleCreateSetNode(value) {
        try {
          const { parentNode } = this.createInfo
          const nextModel = this.topologyModels.find(model => model.bk_obj_id === 'set')
          const formData = (value.sets || []).map(set => ({
            ...set,
            bk_biz_id: this.bizId,
            bk_parent_id: parentNode.data.bk_inst_id
          }))
          const data = await this.createSet(formData)
          const insertBasic = parentNode.data.bk_obj_id === 'biz' ? 1 : 0
          data && data.forEach((set, index) => {
            if (set.data) {
              const nodeData = {
                default: 0,
                child: [],
                bk_obj_name: nextModel.bk_obj_name,
                bk_obj_id: nextModel.bk_obj_id,
                host_count: 0,
                service_instance_count: 0,
                service_template_id: value.service_template_id,
                bk_inst_id: set.data.bk_set_id,
                bk_inst_name: set.data.bk_set_name,
                set_template_id: value.set_template_id,
                status: 'finished'
              }
              this.$refs.tree.addNode(nodeData, parentNode.id, insertBasic + index)
              if (value.set_template_id) {
                this.addModulesInSetTemplate(nodeData, set.data.bk_set_id)
              }
            } else {
              this.$error(set.error_message)
            }
          })
          this.$success(this.$t('新建成功'))
          this.createInfo.show = false
        } catch (e) {
          console.error(e)
        }
      },
      async addModulesInSetTemplate(parentNodeData, id) {
        const modules = await this.$store.dispatch('objectModule/searchModule', {
          bizId: this.bizId,
          setId: id,
          params: { bk_biz_id: this.bizId },
          config: {
            requestId: 'searchModule'
          }
        })
        const parentNodeId = this.getNodeId(parentNodeData)
        const nextModel = this.topologyModels.find(model => model.bk_obj_id === 'module')
        modules.info && modules.info.forEach((_module) => {
          const nodeData = {
            default: 0,
            child: [],
            bk_obj_name: nextModel.bk_obj_name,
            bk_obj_id: nextModel.bk_obj_id,
            host_count: 0,
            service_instance_count: 0,
            service_template_id: _module.service_template_id,
            bk_inst_id: _module.bk_module_id,
            bk_inst_name: _module.bk_module_name,
            status: 'finished'
          }
          this.$refs.tree.addNode(nodeData, parentNodeId, 0)
        })
      },
      async createSet(value) {
        const data = await this.$store.dispatch('objectSet/createset', {
          bizId: this.bizId,
          params: {
            sets: value.map(set => ({
              ...set,
              bk_supplier_account: this.supplierAccount
            }))
          }
        })
        return data || []
      },
      async createModule(value) {
        const data = await this.$store.dispatch('objectModule/createModule', {
          bizId: this.bizId,
          setId: this.createInfo.parentNode.data.bk_inst_id,
          params: {
            ...value,
            bk_biz_id: this.bizId,
            bk_supplier_account: this.supplierAccount
          }
        })
        return {
          bk_inst_id: data.bk_module_id,
          bk_inst_name: data.bk_module_name
        }
      },
      async createCommonInstance(value) {
        const data = await this.$store.dispatch('objectCommonInst/createInst', {
          objId: this.createInfo.nextModelId,
          params: value
        })
        return {
          bk_inst_id: data.bk_inst_id,
          bk_inst_name: data.bk_inst_name
        }
      },
      isTemplate(node) {
        return node.data.service_template_id || node.data.set_template_id
      },
      async refreshCount({ hosts, target }) {
        const nodes = []
        if (target) {
          const node = this.$refs.tree.getNodeById(`${target.data.bk_obj_id}-${target.data.bk_inst_id}`)
          node && nodes.push(node, ...node.parents)
        }
        hosts.forEach(({ module: modules }) => {
          modules.forEach((module) => {
            const node = this.$refs.tree.getNodeById(`module-${module.bk_module_id}`)
            node && nodes.push(node, ...node.parents)
          })
        })
        const nodeSet = new Set()
        const uniqueNodes = nodes.filter((node) => {
          if (nodeSet.has(node)) return false
          nodeSet.add(node)
          return true
        })
        this.setNodeCount(uniqueNodes, true)
      },
      refreshCountByNode(node) {
        const currentNode = node || this.selectedNode
        const nodes = []
        const treeNode = this.$refs.tree.getNodeById(currentNode.id)
        if (treeNode) {
          nodes.push(treeNode, ...treeNode.parents)
        }
        this.setNodeCount(nodes, true)
      },
      handleResize() {
        this.$refs.tree.resize()
      }
    }
  }
</script>

<style lang="scss" scoped>
    .tree-layout {
        overflow: hidden;
    }
    .tree-search {
        display: block;
        width: auto;
        margin: 0 20px;
    }
    .topology-tree {
        padding: 10px 0;
        margin-right: 2px;
        @include scrollbar-y(6px);
        .node-icon {
            display: block;
            width: 20px;
            height: 20px;
            margin: 8px 4px 8px 0;
            border-radius: 50%;
            background-color: #C4C6CC;
            line-height: 1.666667;
            text-align: center;
            font-size: 12px;
            font-style: normal;
            color: #FFF;
            &.is-template {
                background-color: #97aed6;
            }
            &.is-selected {
                background-color: #3A84FF;
            }
        }
        .node-name {
            display: block;
            height: 36px;
            line-height: 36px;
            overflow: hidden;
            @include ellipsis;
        }
        .node-count {
            padding: 0 5px;
            margin: 9px 20px 9px 4px;
            height: 18px;
            line-height: 17px;
            border-radius: 2px;
            background-color: #f0f1f5;
            color: #979ba5;
            font-size: 12px;
            text-align: center;
            &.is-selected {
                background-color: #a2c5fd;
                color: #fff;
            }
            &.loading {
              background-color: transparent;
            }
        }
        .internal-node-icon{
            width: 20px;
            height: 20px;
            line-height: 20px;
            text-align: center;
            margin: 8px 4px 8px 0;
            &.is-selected {
                color: #FFB400;
            }
        }
    }
    .node-info {
        &:hover {
            .info-create-trigger {
                display: inline-block;
                & ~ .node-count {
                    display: none;
                }
            }
        }
        .info-create-trigger {
            display: none;
            font-size: 0;
            &.hovering {
                display: inline-block;
                & ~ .node-count {
                    display: none;
                }
            }
        }
        .node-button {
            height: 24px;
            padding: 0 6px;
            margin: 0 20px 0 4px;
            line-height: 22px;
            border-radius: 4px;
            font-size: 12px;
            min-width: auto;
            &.disabled-node-button {
                @include inlineBlock;
                line-height: 24px;
                font-style: normal;
                background-color: #dcdee5;
                color: #ffffff;
                outline: none;
                cursor: not-allowed;
            }
        }
    }
</style>
