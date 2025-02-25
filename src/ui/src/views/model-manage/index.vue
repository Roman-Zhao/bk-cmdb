<template>
  <div class="group-wrapper"
    v-bkloading="{ isLoading: $loading(requestIds.searchClassifications) }"
    :style="{ 'padding-top': topPadding + 'px' }">
    <cmdb-main-inject ref="mainInject"
      inject-type="prepend"
      :class="['btn-group', 'clearfix', { sticky: !!scrollTop }]">
      <cmdb-tips
        class="mb10"
        tips-key="modelTips"
        more-link="https://bk.tencent.com/docs/markdown/配置平台/产品白皮书/产品功能/Model.md">
        {{$t('模型顶部提示')}}
      </cmdb-tips>
      <div class="fl">
        <cmdb-auth :auth="{ type: $OPERATION.C_MODEL }">
          <bk-button slot-scope="{ disabled }"
            theme="primary"
            :disabled="disabled || modelType === 'disabled'"
            @click="showModelDialog('bk_uncategorized')">
            {{$t('新建模型')}}
          </bk-button>
        </cmdb-auth>
        <cmdb-auth :auth="{ type: $OPERATION.C_MODEL_GROUP }">
          <bk-button slot-scope="{ disabled }"
            theme="default"
            :disabled="disabled || modelType === 'disabled'"
            @click="showGroupDialog(false)">
            {{$t('新建分组')}}
          </bk-button>
        </cmdb-auth>
      </div>
      <div class="model-type-options fr">
        <bk-button class="model-type-button"
          :class="[{ 'model-type-button-active': modelType === '' }]"
          size="small"
          @click="modelType = ''">
          {{$t('全部')}}
        </bk-button>
        <bk-button class="model-type-button"
          :class="[{ 'model-type-button-active': modelType === 'enable' }]"
          size="small"
          @click="modelType = 'enable'">
          {{$t('启用中')}}
        </bk-button>
        <span class="inline-block-middle" style="outline: 0;" v-bk-tooltips="disabledModelBtnText">
          <bk-button class="model-type-button disabled"
            :class="[{ 'model-type-button-active': modelType === 'disabled' }]"
            size="small"
            :disabled="!disabledClassifications.length"
            @click="modelType = 'disabled'">
            {{$t('已停用')}}
          </bk-button>
        </span>
      </div>
      <div class="model-search-options fr">
        <bk-input class="search-model"
          :clearable="true"
          :right-icon="'bk-icon icon-search'"
          :placeholder="$t('请输入关键字')"
          v-model.trim="searchModel">
        </bk-input>
      </div>
    </cmdb-main-inject>

    <ul class="group-list" v-show="currentClassifications.length">
      <li class="group-item clearfix"
        v-for="(classification, classIndex) in currentClassifications"
        :key="classIndex">
        <div
          class="group-title"
          v-bk-tooltips="{
            disabled: !isBuiltinClass(classification),
            content: $t('内置模型组不支持删除和修改'),
            placement: 'right'
          }">
          <div class="title-info">
            <span class="mr5">{{classification['bk_classification_name']}}</span>
            <span class="number">({{classification['bk_objects'].length}})</span>
          </div>
          <template v-if="modelType !== 'disabled'">
            <cmdb-auth class="group-btn ml5"
              :auth="{ type: $OPERATION.C_MODEL, relation: [classification.id] }">
              <bk-button slot-scope="{ disabled }"
                theme="primary"
                text
                :disabled="disabled"
                @click="showModelDialog(classification.bk_classification_id)">
                <i class="icon-cc-add-line"></i>
              </bk-button>
            </cmdb-auth>
            <cmdb-auth
              v-if="!isBuiltinClass(classification)"
              class="group-btn"
              :auth="{ type: $OPERATION.U_MODEL_GROUP, relation: [classification.id] }">
              <bk-button slot-scope="{ disabled }"
                theme="primary"
                text
                :disabled="disabled"
                @click="showGroupDialog(true, classification)">
                <i class="icon-cc-edit"></i>
              </bk-button>
            </cmdb-auth>
            <cmdb-auth
              v-if="!isBuiltinClass(classification)"
              class="group-btn"
              :auth="{ type: $OPERATION.D_MODEL_GROUP, relation: [classification.id] }">
              <bk-button slot-scope="{ disabled }"
                theme="primary"
                text
                :disabled="disabled"
                @click="deleteGroup(classification)">
                <i class="icon-cc-delete"></i>
              </bk-button>
            </cmdb-auth>
          </template>
        </div>
        <ul class="model-list clearfix">
          <li class="model-item bgc-white"
            @mouseenter="handleModelMouseEnterDebounce(model)"
            :class="{
              'ispaused': model['bk_ispaused'],
              'ispre': model.ispre
            }"
            v-for="(model, modelIndex) in classification['bk_objects']"
            :key="modelIndex">
            <div class="info-model"
              :class="{
                'radius': model.bk_ispaused || isNoInstanceModel(model.bk_obj_id)
              }"
              @click="modelClick(model)">
              <div class="icon-box">
                <i class="icon" :class="[model['bk_obj_icon']]"></i>
              </div>
              <div class="model-details">
                <p class="model-name" :title="model['bk_obj_name']">{{model['bk_obj_name']}}</p>
                <p class="model-id" :title="model['bk_obj_id']">{{model['bk_obj_id']}}</p>
              </div>
            </div>
            <div v-if="!model.bk_ispaused && !isNoInstanceModel(model.bk_obj_id)"
              class="info-instance"
              @click="handleGoInstance(model)">
              <i class="icon-cc-share"></i>
              <p>
                <cmdb-loading
                  :loading="!modelStatisticsSet[model.bk_obj_id] || $loading(requestIds.statistics[model.bk_obj_id])">
                  {{modelStatisticsSet[model.bk_obj_id] | instanceCount}}
                </cmdb-loading>
              </p>
            </div>
          </li>
        </ul>
      </li>
    </ul>
    <no-search-results v-if="!currentClassifications.length" :text="$t('搜不到相关模型')" />

    <bk-dialog
      class="bk-dialog-no-padding bk-dialog-no-tools group-dialog dialog"
      :close-icon="false"
      :width="600"
      :mask-close="false"
      v-model="groupDialog.isShow">
      <div class="dialog-content">
        <p class="title">{{groupDialog.title}}</p>
        <div class="content">
          <label>
            <div class="label-title">
              {{$t('唯一标识')}}<span class="color-danger">*</span>
            </div>
            <div class="cmdb-form-item" :class="{ 'is-error': errors.has('classifyId') }">
              <bk-input type="text" class="cmdb-form-input"
                name="classifyId"
                :placeholder="$t('请输入唯一标识')"
                :disabled="groupDialog.isEdit"
                v-model.trim="groupDialog.data['bk_classification_id']"
                v-validate="'required|classifyId|length:128|reservedWord'">
              </bk-input>
              <p class="form-error" :title="errors.first('classifyId')">{{errors.first('classifyId')}}</p>
            </div>
            <i class="bk-icon icon-info-circle" v-bk-tooltips="$t('请填写英文开头，下划线，数字，英文的组合')"></i>
          </label>
          <label>
            <span class="label-title">
              {{$t('名称')}}
            </span>
            <span class="color-danger">*</span>
            <div class="cmdb-form-item" :class="{ 'is-error': errors.has('classifyName') }">
              <bk-input type="text"
                class="cmdb-form-input"
                name="classifyName"
                :placeholder="$t('请输入名称')"
                v-model.trim="groupDialog.data['bk_classification_name']"
                v-validate="'required|length:128'">
              </bk-input>
              <p class="form-error" :title="errors.first('classifyName')">{{errors.first('classifyName')}}</p>
            </div>
          </label>
        </div>
      </div>
      <div slot="footer" class="footer">
        <bk-button theme="primary"
          :loading="$loading(['updateClassification', 'createClassification'])"
          @click="saveGroup">
          {{groupDialog.isEdit ? $t('保存') : $t('提交')}}
        </bk-button>
        <bk-button theme="default" @click="hideGroupDialog">{{$t('取消')}}</bk-button>
      </div>
    </bk-dialog>

    <the-create-model
      :is-show.sync="modelDialog.isShow"
      :group-id.sync="modelDialog.groupId"
      :title="$t('新建模型')"
      :operating="$loading('createModel')"
      @confirm="saveModel">
    </the-create-model>

    <bk-dialog
      class="bk-dialog-no-padding"
      :width="400"
      :show-footer="false"
      :mask-close="false"
      v-model="sucessDialog.isShow">
      <div class="success-content">
        <i class="bk-icon icon-check-1"></i>
        <p>{{$t('模型创建成功')}}</p>
        <div class="btn-box">
          <bk-button theme="primary" @click="modelClick(curCreateModel)">{{$t('配置字段')}}</bk-button>
          <bk-button @click="sucessDialog.isShow = false">{{$t('返回列表')}}</bk-button>
        </div>
      </div>
    </bk-dialog>
  </div>
</template>

<script>
  import has from 'has'
  import cmdbMainInject from '@/components/layout/main-inject'
  import theCreateModel from '@/components/model-manage/_create-model'
  import cmdbLoading from '@/components/loading/index.vue'
  import noSearchResults from '@/views/status/no-search-results.vue'
  import { mapGetters, mapMutations, mapActions } from 'vuex'
  import debounce from 'lodash.debounce'
  import { addMainScrollListener, removeMainScrollListener } from '@/utils/main-scroller'
  import { addResizeListener, removeResizeListener } from '@/utils/resize-events'
  import {
    MENU_RESOURCE_INSTANCE,
    MENU_MODEL_DETAILS
  } from '@/dictionary/menu-symbol'
  import { BUILTIN_MODEL_RESOURCE_MENUS } from '@/dictionary/model-constants.js'

  export default {
    filters: {
      instanceCount(value) {
        if (!value) return
        if (value?.error) {
          return '--'
        }
        return value.inst_count > 999 ? '999+' : value.inst_count
      }
    },
    components: {
      theCreateModel,
      cmdbMainInject,
      noSearchResults,
      cmdbLoading
    },
    data() {
      return {
        scrollHandler: null,
        scrollTop: 0,
        topPadding: 0,
        modelType: '',
        searchModel: '',
        filterClassifications: [],
        modelStatisticsSet: {},
        curCreateModel: {},
        sucessDialog: {
          isShow: false
        },
        groupDialog: {
          isShow: false,
          isEdit: false,
          title: this.$t('新建分组'),
          data: {
            bk_classification_id: '',
            bk_classification_name: '',
            id: ''
          }
        },
        modelDialog: {
          isShow: false,
          groupId: ''
        },
        requestIds: {
          statistics: [], // 模型实例数据加载请求 id
          searchClassifications: Symbol('searchClassifications') // 模型分组数据加载请求 id
        },
        mainlineModels: []
      }
    },
    computed: {
      ...mapGetters(['supplierAccount', 'userName']),
      ...mapGetters('objectModelClassify', [
        'classifications'
      ]),
      allClassifications() {
        const allClassifications = []
        this.classifications.forEach((classification) => {
          allClassifications.push({
            ...classification,
            bk_objects: classification.bk_objects
              .filter(model => !model.bk_ishidden)
              .sort((a, b) => a.bk_ispaused - b.bk_ispaused),
          })
        })
        return allClassifications
      },
      enableClassifications() {
        const enableClassifications = []
        this.allClassifications.forEach((classification) => {
          enableClassifications.push({
            ...classification,
            bk_objects: classification.bk_objects.filter(model => !model.bk_ispaused),
          })
        })
        return enableClassifications.filter(item => item.bk_objects.length)
      },
      disabledClassifications() {
        const disabledClassifications = []
        this.classifications.forEach((classification) => {
          disabledClassifications.push({
            ...classification,
            bk_objects: classification.bk_objects.filter(model => model.bk_ispaused),
          })
        })
        return disabledClassifications.filter(item => item.bk_objects.length)
      },
      currentClassifications() {
        let currentClassifications = []

        if (!this.searchModel && !this.modelType) {
          currentClassifications = this.allClassifications
        }

        if (this.modelType) {
          currentClassifications = this.modelType === 'enable' ? this.enableClassifications : this.disabledClassifications
        }

        if (this.searchModel) {
          currentClassifications = this.filterClassifications
        }

        return currentClassifications.sort((a, b) => (b.bk_classification_id === 'bk_uncategorized' ? -1 : 0))
      },
      disabledModelBtnText() {
        return this.disabledClassifications.length ? '' : this.$t('停用模型提示')
      }
    },
    watch: {
      searchModel(value) {
        if (!value) {
          return
        }
        const searchResult = []
        // eslint-disable-next-line no-nested-ternary
        const currentClassifications = !this.modelType ? this.allClassifications : (this.modelType === 'enable' ? this.enableClassifications : this.disabledClassifications)
        const classifications = this.$tools.clone(currentClassifications)
        const lowerCaseValue = value.toLowerCase()
        for (let i = 0; i < classifications.length; i++) {
          classifications[i].bk_objects = classifications[i].bk_objects.filter((model) => {
            const modelName = model.bk_obj_name.toLowerCase()
            const modelId = model.bk_obj_id.toLowerCase()
            // eslint-disable-next-line max-len
            return (modelName && modelName.indexOf(lowerCaseValue) !== -1) || (modelId && modelId.indexOf(lowerCaseValue) !== -1)
          })
          searchResult.push(classifications[i])
        }
        this.filterClassifications = searchResult.filter(item => item.bk_objects.length)
      },
      modelType() {
        this.searchModel = ''
      }
    },
    async created() {
      this.scrollHandler = (event) => {
        this.scrollTop = event.target.scrollTop
      }

      addMainScrollListener(this.scrollHandler)

      this.handleModelMouseEnterDebounce = debounce(this.handleModelItemMouseEnter, 200)

      try {
        await this.searchClassificationsObjects({
          params: {},
          config: {
            requestId: this.requestIds.searchClassifications
          }
        })
      } catch (e) {
        this.$route.meta.view = 'error'
      }

      try {
        this.mainlineModels = await this.searchMainlineObject({})
      } catch (error) {
        console.log(error)
      }

      if (this.$route.query.searchModel) {
        const { hash } = window.location
        this.searchModel = this.$route.query.searchModel
        window.location.hash = hash.substring(0, hash.indexOf('?'))
      }
    },
    mounted() {
      addResizeListener(this.$refs.mainInject.$el, this.handleSetPadding)
    },
    beforeDestroy() {
      removeResizeListener(this.$refs.mainInject.$el, this.handleSetPadding)
      removeMainScrollListener(this.scrollHandler)
      this.$http.cancelRequest(this.requestIds.statistics)
    },
    methods: {
      ...mapMutations('objectModelClassify', [
        'updateClassify',
        'deleteClassify'
      ]),
      ...mapActions('objectModelClassify', [
        'searchClassificationsObjects',
        'getClassificationsObjectStatistics',
        'createClassification',
        'updateClassification',
        'deleteClassification'
      ]),
      ...mapActions('objectModel', [
        'createObject'
      ]),
      ...mapActions('objectMainLineModule', [
        'searchMainlineObject',
      ]),
      handleSetPadding() {
        this.topPadding = this.$refs.mainInject.$el.offsetHeight
      },
      isBuiltinClass(classification) {
        return classification.bk_classification_type === 'inner'
      },
      showGroupDialog(isEdit, group) {
        if (isEdit) {
          this.groupDialog.data.id = group.id
          this.groupDialog.title = this.$t('编辑分组')
          this.groupDialog.data.bk_classification_id = group.bk_classification_id
          this.groupDialog.data.bk_classification_name = group.bk_classification_name
          this.groupDialog.data.id = group.id
        } else {
          this.groupDialog.title = this.$t('新建分组')
          this.groupDialog.data.bk_classification_id = ''
          this.groupDialog.data.bk_classification_name = ''
          this.groupDialog.data.id = ''
        }
        this.groupDialog.isEdit = isEdit
        this.groupDialog.isShow = true
      },
      hideGroupDialog() {
        this.groupDialog.isShow = false
        this.$validator.reset()
      },
      async getModelInstanceCount(id) {
        // 存在则不再请求
        if (has(this.modelStatisticsSet, id)) {
          return
        }

        const requestId = `getModelInstanceCount_${id}`
        this.requestIds.statistics.push(requestId)

        // 取消上一个请求
        const currentIndex = this.requestIds.statistics.findIndex(rid => rid === requestId)
        const prevIndex = this.requestIds.statistics[currentIndex - 1]
        if (prevIndex) {
          this.$http.cancelRequest(prevIndex)
        }

        const result = await this.$store.dispatch('objectCommonInst/searchInstanceCount', {
          params: {
            condition: { obj_ids: [id] }
          },
          config: {
            requestId,
            globalError: false
          }
        })

        const [data] = result
        this.$set(this.modelStatisticsSet, data.bk_obj_id, {
          error: data.error,
          inst_count: data.inst_count
        })
      },
      async saveGroup() {
        try {
          const res = await Promise.all([
            this.$validator.validate('classifyId'),
            this.$validator.validate('classifyName')
          ])
          if (res.includes(false)) {
            return
          }
          const params = {
            bk_supplier_account: this.supplierAccount,
            bk_classification_id: this.groupDialog.data.bk_classification_id,
            bk_classification_name: this.groupDialog.data.bk_classification_name
          }
          if (this.groupDialog.isEdit) {
            // eslint-disable-next-line no-unused-vars
            const res = await this.updateClassification({
              id: this.groupDialog.data.id,
              params,
              config: {
                requestId: 'updateClassification'
              }
            })
            this.updateClassify({ ...params, ...{ id: this.groupDialog.data.id } })
          } else {
            const res = await this.createClassification({
              params,
              config: { requestId: 'createClassification' }
            })
            this.updateClassify({ ...params, ...{ id: res.id } })
            this.$success(this.$t('新建成功'))
          }
          this.hideGroupDialog()
          this.searchModel = ''
        } catch (error) {
          console.log(error)
        }
      },
      deleteGroup(group) {
        this.$bkInfo({
          title: this.$t('确认要删除此分组'),
          confirmFn: async () => {
            try {
              await this.deleteClassification({
                id: group.id
              })
              this.$store.commit('objectModelClassify/deleteClassify', group.bk_classification_id)
              this.searchModel = ''
              this.$success(this.$t('删除成功'))
            } catch (error) {
              console.log(error)
            }
          }
        })
      },
      showModelDialog(groupId) {
        this.modelDialog.groupId = groupId || ''
        this.modelDialog.isShow = true
      },
      async saveModel(data) {
        const params = {
          bk_supplier_account: this.supplierAccount,
          bk_obj_name: data.bk_obj_name,
          bk_obj_icon: data.bk_obj_icon,
          bk_classification_id: data.bk_classification_id,
          bk_obj_id: data.bk_obj_id,
          userName: this.userName
        }
        try {
          const createModel = await this.createObject({ params, config: { requestId: 'createModel' } })
          this.curCreateModel = createModel
          this.sucessDialog.isShow = true
          this.$http.cancel('post_searchClassificationsObjects')
          this.getModelInstanceCount(params.bk_obj_id)
          this.searchClassificationsObjects({
            params: {}
          })
          this.modelDialog.isShow = false
          this.modelDialog.groupId = ''
          this.searchModel = ''
        } catch (error) {
          console.log(error)
        }
      },
      modelClick(model) {
        this.$store.commit('objectModel/setActiveModel', model)
        this.$routerActions.redirect({
          name: MENU_MODEL_DETAILS,
          params: {
            modelId: model.bk_obj_id
          },
          history: true
        })
      },
      handleGoInstance(model) {
        this.sucessDialog.isShow = false
        if (has(BUILTIN_MODEL_RESOURCE_MENUS, model.bk_obj_id)) {
          const query = model.bk_obj_id === 'host' ? { scope: 'all' } : {}
          this.$routerActions.redirect({
            name: BUILTIN_MODEL_RESOURCE_MENUS[model.bk_obj_id],
            query
          })
        } else {
          this.$routerActions.redirect({
            name: MENU_RESOURCE_INSTANCE,
            params: {
              objId: model.bk_obj_id
            }
          })
        }
      },
      isNoInstanceModel(modelId) {
        // 不能直接查看实例的模型
        const noInstanceModelIds = ['set', 'module']
        return noInstanceModelIds.includes(modelId)
      },
      handleModelItemMouseEnter(model) {
        if (!model) return

        const isDisabledModel = model.bk_ispaused && !model.bk_ishidden

        if (isDisabledModel || this.isNoInstanceModel(model.bk_obj_id)) return

        this.getModelInstanceCount(model.bk_obj_id)
      }
    }
  }
</script>

<style lang="scss" scoped>
    .group-wrapper {
        padding: 72px 0 20px 0;
    }
    .btn-group {
        position: absolute;
        top: 53px;
        left: 0;
        width: calc(100% - 17px);
        padding: 15px 20px 20px;
        font-size: 0;
        background-color: #fafbfd;
        z-index: 100;
        .bk-primary {
            margin-right: 10px;
        }
        &.sticky {
            box-shadow: 0 0 8px 1px rgba(0, 0, 0, 0.03);
        }
    }
    .model-search-options {
        .search-model {
            width: 240px;
        }
    }
    .model-type-options {
        margin: 0 0 0 10px;
        font-size: 0;
        text-align: right;
        .model-type-button {
            position: relative;
            margin: 0;
            font-size: 12px;
            height: 32px;
            line-height: 30px;
            &.enable {
                border-radius: 2px 0 0 2px;
                border-right-color: #3a84ff;
                z-index: 2;
            }
            &.disabled {
                border-radius: 0 2px 2px 0;
                margin-left: -1px;
                z-index: 1;
            }
            &:hover {
                border-color: #3a84ff;
                z-index: 2;
            }
            &-active {
                border-color: #3a84ff;
                color: #3a84ff;
                z-index: 2;
            }
            & + .model-type-button {
              border-radius: 0 2px 2px 0;
              margin-left: -1px;
            }
        }
    }
    .group-list {
        padding: 0 20px;
        .group-item {
            position: relative;
            padding: 10px 0 20px;
        }
        .group-title {
            display: inline-block;
            margin: 0 40px 0 0;
            height: 22px;
            line-height: 22px;
            color: #333948;
            outline: 0;
            &:before {
                content: "";
                display: inline-block;
                width:4px;
                height:14px;
                margin: 0 10px 0 0;
                vertical-align: middle;
                background: $cmdbBorderColor;
            }
            .title-info {
                @include inlineBlock;
                font-size: 0;
                > span {
                    @include inlineBlock;
                    font-size: 14px;
                    font-weight: 700;
                }
            }
            .number {
                color: $cmdbBorderColor;
            }
            .group-btn {
                display: none;
                vertical-align: middle;
                margin-right: 4px;
                .bk-button-text {
                    font-size: 16px;
                }
            }
            &:hover {
                .group-btn {
                    display: inline-block;
                }
            }
        }
    }
    .model-list {
        padding-left: 12px;
        overflow: hidden;
        transition: height .2s;
        .model-item {
            display: flex;
            position: relative;
            float: left;
            margin: 10px 10px 0 0;
            width: calc((100% - 10px * 4) / 5);
            height: 70px;
            border: 1px solid $cmdbTableBorderColor;
            border-radius: 4px;
            background-color: #ffffff;
            cursor: pointer;
            &:nth-child(5n) {
                margin-right: 0;
            }
            &.ispaused {
                background: #fcfdfe;
                border-color: #dde4eb;
                .icon-box {
                    color: #96c2f7;
                }
                .model-name {
                    color: #bfc7d2;
                }
            }
            &.ispre {
                .icon-box {
                    color: #798aad;
                }
            }
            &:hover {
                border-color: $cmdbBorderFocusColor;
                .info-instance {
                    display: block;
                }
            }
            .icon-box {
                float: left;
                width: 66px;
                text-align: center;
                font-size: 32px;
                color: #3a84ff;
                .icon {
                    line-height: 68px;
                }
            }
            .model-details {
                padding: 0 4px 0 0;
                overflow: hidden;
            }
            .model-name {
                margin-top: 16px;
                line-height: 19px;
                font-size: 14px;
                @include ellipsis;
            }
            .model-id {
                line-height: 16px;
                font-size: 12px;
                color: #bfc7d2;
                @include ellipsis;
            }
            .info-model {
                flex: 1;
                width: 0;
                border-radius: 4px 0 0 4px;
                &:hover {
                    background-color: #f0f5ff;
                }
                &.radius {
                    border-radius: 4px;
                }
            }
            .info-instance {
                display: none;
                width: 70px;
                padding: 0 8px 0 6px;
                text-align: center;
                color: #c3cdd7;
                border-radius: 0 4px 4px 0;
                &:hover {
                    background-color: #f0f5ff;
                    p {
                        color: #3a84ff;
                    }
                }
                .icon-cc-share {
                    font-size: 14px;
                    margin-top: 16px;
                    color: #3a84ff;
                }
                p {
                    font-size: 16px;
                    padding-top: 2px;
                    @include ellipsis;
                }
            }
        }
    }
    .dialog {
        .dialog-content {
            padding: 20px 15px 20px 28px;
        }
        .title {
            font-size: 20px;
            color: #333948;
            line-height: 1;
            padding-bottom: 14px;
        }
        .label-item,
        label {
            display: block;
            margin-bottom: 10px;
            font-size: 0;
            &:last-child {
                margin: 0;
            }
            .color-danger {
                display: inline-block;
                font-size: 16px;
                width: 15px;
                text-align: center;
                vertical-align: middle;
            }
            .icon-info-circle {
                font-size: 18px;
                color: $cmdbBorderColor;
            }
            .label-title {
                font-size: 16px;
                line-height: 36px;
                vertical-align: middle;
                @include ellipsis;
            }
            .cmdb-form-item {
                display: inline-block;
                margin-right: 10px;
                width: 519px;
                vertical-align: middle;
            }
        }
        .footer {
            font-size: 0;
            text-align: right;
            .bk-primary {
                margin-right: 10px;
            }
        }
    }
    .success-content {
        text-align: center;
        padding-bottom: 46px;
        p {
            color: #444444;
            font-size: 24px;
            padding: 10px 0 20px;
        }
        .icon-check-1 {
            width: 58px;
            height: 58px;
            line-height: 58px;
            font-size: 50px;
            font-weight: bold;
            color: #fff;
            border-radius: 50%;
            background-color: #2dcb56;
            text-align: center;
        }
        .btn-box {
            font-size: 0;
            .bk-button {
                margin: 0 5px;
            }
        }
    }
</style>
