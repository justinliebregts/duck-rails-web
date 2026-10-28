<script>
import { CModal, CRow } from '@coreui/vue';

</script>
<template>
  <div>
    <CContainer>
      <CHeader class="mb-3 p-3">
        <h1>Meta Attestation Account: {{ pageTitle }}</h1>
      </CHeader>

      <CCard class="mb-4">
        <CCardHeader>
          <h3>Account Identifiers</h3>
        </CCardHeader>
        <CCardBody v-if="metaInfo !== null">
          <CRow class="mb-2">
            <CCol sm="4" class="font-weight-bold">Alias:</CCol>
            <CCol sm="8">{{ metaInfo.alias }}</CCol>
          </CRow>
          <CRow class="mb-2">
            <CCol sm="4" class="font-weight-bold">App-Scoped ID (for each app):</CCol>
            <CCol sm="8">
              <strong><code>{{ metaInfo.id }}</code></strong>
            </CCol>
          </CRow>
          <CRow class="mb-2">
            <CCol sm="4" class="font-weight-bold">Org-Scoped ID (the same in the Org):</CCol>
            <CCol sm="8">{{ metaInfo.scoped_id }}</CCol>
          </CRow>
          <CRow>
            <CCol sm="12">
              <br/>
            </CCol>
          </CRow >

          <CRow v-if="banForm !== null && banForm.unique_id !== null">
            <CCol sm="12">
              <h4>Remote Attestation Status</h4>
            </CCol>

            <CCol sm="4" class="font-weight-bold">Assigned Unique ID:</CCol>
            <CCol sm="8">{{ banForm.unique_id }}</CCol>
          
            <CCol sm="4" class="font-weight-bold">Was Banned:</CCol>
            <CCol sm="8">
              <CBadge v-if="remote_status.is_banned" color="danger" class="px-2 py-1 fs-6 fw-bold">
                  TRUE
              </CBadge>
              <CBadge v-else color="secondary">
                  FALSE
              </CBadge>
            </CCol>

            <CCol sm="4" class="font-weight-bold">Remaining Time:</CCol>
            <CCol sm="8">{{ remote_status.remaining_ban_time || 'N/A'  }}</CCol>

            <CCol sm="4" class="font-weight-bold">Ban Creation Date:</CCol>
            <CCol sm="8">{{ remote_status.ban_creation_date || 'N/A' }}</CCol>
            
            <CCol sm="12">
              <CButton color="primary" @click="showBanModal = true">
                Toggle Attestation Ban
              </CButton>
            </CCol>
          </CRow>
          <CRow>
          </CRow>
        </CCardBody>
      </CCard>

      <!--template v-if="!paginationEnabled">
        <div style="margin-bottom: 10px;"><b>Pages:</b> {{ numItemPages }}</div>
        <CButtonGroup class="mb-3">
          <CButton @click="pageTable(-10)" color="dark"> &#60;&#60;&#60; </CButton>&nbsp;
          <CButton @click="pageTable(-1)" color="dark">&#60;</CButton>&nbsp;
          <CButton @click="pageTable(1)" color="dark">&#62;</CButton>&nbsp;
          <CButton @click="pageTable(10)" color="dark">&#62;&#62;&#62;</CButton>
          <CInput id="page-input" :value.sync="activePage"/>
          <CCol>
          </CCol>

        </CButtonGroup>
      </template-->
      
      <AttestationRecordsTable
        :records="raw_attestation_records"
        :pagination-enabled="paginationEnabled"                    
        @select-record="selectRecordInfo"
      />
      
      <AttestationBanRecordsTable
        :records="raw_attestation_ban_records"
        :pagination-enabled="paginationEnabled"
        @select-ban="selectBanInfo"
      />  

      <CToaster :autohide="2000">
        <template v-for="toast in fixedToasts">
          <CToast
              :key="'toast' + toast"
              :show="true"
              :color="toastColor[toast]"
          >
            {{ toastMsg[toast] }}
          </CToast>
        </template>
      </CToaster>

      
  <CModal 
  :show.sync="showBanModal"
  title="Bans with Attestation"
  @update:show="onModalClose"
  >
  <CForm @submit.prevent="submitDeviceBan">
    <CRow class="align-items-center">
      <!-- Checkbox for is_banning status -->
      <CCol md="12" class="mb-3 d-flex align-items-center pt-3">
        <CInputCheckbox
          id="is_banning"
          label="Active Ban (is_banning)"
          v-model="banForm.is_banning"
          @update:checked="(val) => $set(banForm, 'is_banning', val)"          
        />
      </CCol>

      <!-- Unique Device ID -->
      <CCol md="6" class="mb-3">
        <CInput
          label="Device Unique ID"
          type="text"
          id="unique_id"
          placeholder="Enter unique_id"
          v-model="banForm.unique_id"
          readonly
        />
      </CCol>

      <!-- Ban ID -->
      <CCol md="6" class="mb-3">
        <CInput
          label="Ban ID (Optional)"
          type="text"
          id="ban_id"
          placeholder="Leave blank for new ban"
          v-model="banForm.ban_id"
        />
      </CCol>
    </CRow>

    <CRow class="align-items-center">
      <!-- Duration in Minutes -->
      <CCol md="4" class="mb-3">
        <CInput
          label="Remaining Time (Minutes)"
          type="number"
          id="remaining_time_in_minute"
          min="1"
          v-model.number="banForm.remaining_time_in_minute"
          :disabled="!banForm.is_banning"
          :required="banForm.is_banning"
        />
      </CCol>

      <!-- Ban Reason -->
      <CCol md="5" class="mb-3">
        <CInput
          label="Ban Reason"
          type="text"
          id="reason"
          placeholder="Reason for ban action"
          v-model="banForm.reason"
        />
      </CCol>
    </CRow>

    <!-- Action Button -->
    <div class="d-flex justify-content-end mt-2">
      <CButton
        type="submit"
        :color="Boolean(banForm.is_banning) ? 'danger' : 'warning'"
        :disabled="submitting"
        @click="submitDeviceBan(); showBanModal=false;"
      >
        {{ Boolean(banForm.is_banning) ? 'Execute Device Ban' : 'Lift Device Ban' }}
      </CButton>
    </div>
  </CForm>

  <template #footer>
    <CButton color="secondary" @click="showBanModal=false;">
      Cancel
    </CButton>
  </template>
</CModal>

    </CContainer>
  </div>

</template>

      

<script>

import AttestationRecordsTable from './AttestationRecordsTable.vue';
import AttestationBanRecordsTable from './AttestationBanRecordsTable.vue';
export default {
  props: ['platform_id'], // Captured from the URL
  computed: {
    isViewMode() {
      return !!this.platform_id; // If ID exists, we are viewing
    },
    pageTitle() {
      return this.isViewMode ? this.platform_id : 'UnKnown';
    },
  },
  components: {
    AttestationRecordsTable,
    AttestationBanRecordsTable,
  },
  data() {
    return {
      toastMsg: [],
      toastColor: [],
      fixedToasts: 0,
      showSingleRecordModal:false,
      paginationEnabled: false,
      //activePage: 1,
      //activeBanListPage: 1,
      //numItemPages: 0,
      //numBanListPages: 0,
      raw_attestation_records:[],
      raw_attestation_ban_records:[],
      showBanModal: false,
      raw_status: null,

      banForm:{
        unique_id:null,
        ban_id:null,
        remaining_time_in_minute:1440,
        reason:null,
        is_banning:null
      },

      remote_status:{
        is_banned: null,
        remaining_time_in_minute: null,
        ban_creation_date: null
      },
      submitting: false,

      metaInfo: null,

    }
  },
  watch: {
    // Triggers automatically whenever `metaInfo` is set or updated
    metaInfo: {
      handler(newVal) {
        if (newVal && newVal.id) {          
          let prefab_id = this.$gstate.getPrefabID();
          this.getUniqueIdsFromElastic(prefab_id, this.platform_id);
        }
      },
      deep: true // Optional: checks nested property changes
    },
    "banForm.unique_id": {
      handler(newVal) {
        if (newVal && newVal.length > 0) {
          let prefab_id = this.$gstate.getPrefabID();
          // console.log("prefab_id", prefab_id);
          // console.log("newVal", newVal);
          this.getRemoteBanStatus(prefab_id, newVal);
        }
      },
      deep: true // Optional: checks nested property changes
    }
  },
  created() {
    this.getMetaScopedinfo();

    let prefab_id = this.$gstate.getPrefabID();
    this.getDeviceBansFromElastic(prefab_id);
  },
  methods: {
    async getMetaScopedinfo() {
      let urlstr = this.$gstate.defaultUrl + '/player/v2/admin/oculus-org-scoped-id'; // ?id=8958627047567860&app_source=true

      let q = { oculus_id: this.platform_id };
      let prefab_id = this.$gstate.getPrefabID();
      if (prefab_id.length > 0)
      {
        q.app_source = prefab_id;
      }

    try {
        const response = await this.axios.get(urlstr, this.$auth.queryWithJsonOptions(q));        
        // Successfully assign data
        this.metaInfo = response.data.data || response.data;
      } catch (error) {
        console.error("Failed to fetch Meta scoped info:", error);

        // Safely extract validation error, server message, or fall back to generic error
        const errData = error.response?.data;
        const errMsg = errData?.messages?.[0] || errData?.message || error.message || "Network Error";

        this.toast(errMsg, "danger");
      } finally {
        // Executes regardless of success or failure (e.g. stop loading spinners)
        // this.loading = false;
      }
    },
    async getUniqueIdsFromElastic(app_source,app_scoped_id) {
      let urlstr = this.$gstate.defaultUrl + '/player/v2/admin/oculus-attestation/list'; // ?id=8958627047567860&app_source=true

       let q = { usr_id: app_scoped_id };

      try{
        const response = await this.axios.get(urlstr, this.$auth.queryWithJsonOptions(q));
        //console.log(response.data.data.records)
        // Successfully assign data
        this.raw_attestation_records = response.data.data.records;
      }catch (error) {
        console.error("Failed to fetch Meta scoped info:", error);

        // Safely extract validation error, server message, or fall back to generic error
        const errData = error.response?.data;
        const errMsg = errData?.messages?.[0] || errData?.message || error.message || "Network Error";

        this.toast(errMsg, "danger");
      } finally {
        // Executes regardless of success or failure (e.g. stop loading spinners)
        // this.loading = false;
      }
    },
    async getDeviceBansFromElastic(app_source) {
      let urlstr = `${this.$gstate.defaultUrl}/player/v2/admin/oculus-attestation/ban-list`;

       let q = { oculus_id: this.platform_id };

      try{
        const response = await this.axios.get(urlstr, this.$auth.queryWithJsonOptions(q));
        //console.log("ban records", response.data.data.records)
        // Successfully assign data
        this.raw_attestation_ban_records = response.data.data.records;
      }catch (error) {
        console.error("Failed to fetch Meta scoped info:", error);

        // Safely extract validation error, server message, or fall back to generic error
        const errData = error.response?.data;
        const errMsg = errData?.messages?.[0] || errData?.message || error.message || "Network Error";

        this.toast(errMsg, "danger");
      } finally {
        // Executes regardless of success or failure (e.g. stop loading spinners)
        // this.loading = false;
      }
    },
    async getRemoteBanStatus(app_source, unique_id) {

      let urlstr = `${this.$gstate.defaultUrl}/player/v2/admin/oculus-attestation/remote-ban-status/${unique_id}`;

      this.axios.get(urlstr, this.$auth.apiJsonOptions)
          .then(response => {
            //this.toast(response.data.message, "success");
            this.remote_status = response.data.data;
          })
          .catch(error => {
            console.log(error)
            if(error.response !== undefined && error.response.data !== undefined) {
              if(error.response.data.messages[0] !== undefined) {
                this.toast(error.response.data.messages[0], "danger");
              }else if(error.response.data.message !== undefined){
                this.toast(error.response.data.message, "danger");
              }
            }
            //this.toast(error, "danger");
          })
          .finally(() => {
            //console.log("final")
          })
    },
    async submitDeviceBan(){
      const urlstr = `${this.$gstate.defaultUrl}/player/v2/admin/oculus-attestation/toggle-ban`;
      this.submitting = true;

      let prefab_id = this.$gstate.getPrefabID();

      try {
        // Construct final payload
        const payload = { ...this.banForm };
        console.log(payload);
        const response = await this.axios.post(urlstr, payload, this.$auth.apiJsonOptions);

        //// Meta returns a permanent `ban_id` upon successful ban creation
        // if (response.data?.ban_id) {
        //   this.banForm.ban_id = response.data.ban_id;
        // }

        this.toast(response.data?.message || "Device ban action processed successfully!", "success");
        await this.getRemoteBanStatus(prefab_id, this.banForm.unique_id);
      } catch (error) {
        console.error("Device ban error:", error);
        
        // Extract server validation error[cite: 1]
        const errData = error.response?.data;
        const errMsg = errData?.messages?.[0] || errData?.message || error.message || "Failed to process ban";
        this.toast(errMsg, "danger");
      } finally {
        this.submitting = false;
      }
    },



    selectRecordInfo: function(item, index, columnName, event){
      //console.log(item);
      //console.log(index);
      //console.log(columnName);
      //console.log(event);

      //this.banForm.remaining_time_in_minute = 1440;
      this.banForm.unique_id = item.unique_id;
      //this.banForm.is_banning = true;
      // this.banForm.ban_id  = "";

      //POP up modal about the claim
      //this.claim_to_Show = item.claims;
    },
    onModalClose(){
      //this.claim_to_Show = null;

      this.banForm={ unique_id:null,ban_id:null,remaining_time_in_minute:1440,reason:null,is_banning:null};
    },

    selectBanInfo: function(item, index, columnName, event){
      //console.log(item);
      //console.log(index);
      //console.log(columnName);
      //console.log(event);

      this.banForm.unique_id = item.unique_id;
      this.banForm.ban_id = item.ban_id;
      this.banForm.is_banning = false;
    },

    // pageTable: function(pagesToPage){
    //   this.activePage = this.activePage + parseInt(pagesToPage);

    //   if(this.activePage < 1){
    //     this.activePage = 1
    //   } else if (this.activePage > this.numItemPages) {
    //     this.activePage = this.numItemPages;
    //   }
    // },
    // onPagesChange: function(numPages){
    //   this.numItemPages = numPages;
    // },

    // onBanListPagesChange: function(numPages){
    //   this.numBanListPages = numPages;
    // },
    toast: function(msg, color = "primary"){
      this.fixedToasts++;
      this.toastMsg[this.fixedToasts] = msg;
      this.toastColor[this.fixedToasts] = color;

    },
  }
}
</script>