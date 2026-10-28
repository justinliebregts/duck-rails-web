<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3 h1">
                IP Block
            </CHeader>
            <CRow>
                <CCol sm="12">
                    <CCard>
                        <CCardHeader class="h3"> CREATE-UPDATE IP Ban </CCardHeader>
                        <CCardBody>
                            <CRow class="align-items-end">
                                <CCol md="5">
                                    <CInput 
                                    label="IP Address"
                                    type="text"
                                    id="ip"
                                    placeholder="Enter IP Address"
                                    v-model="baninfo.ip"
                                    />
                                </CCol>
                                <CCol md="5">
                                    <CInput
                                    label="Reason"
                                    type="text"
                                    id="reason"
                                    placeholder="Enter Reason"
                                    v-model="baninfo.reason"
                                    />
                                </CCol>
                                <CCol md="2" class="mb-3">
                                    <CButton block color="primary" @click="openCreateConfirmModal">
                                    Create/Update
                                    </CButton>
                                </CCol>
                                <CCol sm="12">
                                    <CCard>
                                        <CCardHeader class="h3"> IP Blacklist</CCardHeader>
                                        <CCardBody>
                                    <CDataTable
                                        :items="blacklistParsed"
                                        :fields="['IP', 'Reason', 'CreatedAt', 'UpdatedAt', 'DeletedAt']"
                                        :items-per-page="10"
                                        :hover="true"
                                        @row-clicked="openExecuteConfirmModal"
                                    >
                                    </CDataTable>
                                    </CCardBody>
                                    </CCard>
                                </CCol>
                            </CRow>

                            <CRow>
                                <CCol sm="12">
                                    <CCard>
                                        <CCardHeader class="h3"> Cloudflare IP Rules</CCardHeader>
                                        <CCardBody>
                                            <CRow>
                                                <CCol sm="3">
                                                </CCol>
                                                <CCol sm="5">
                                                </CCol>
                                                <CCol  sm="1">
                                                </CCol>
                                                <CCol v-if="cloudflareBlacklist != null && cloudflareBlacklist.length > 0" sm="3" class="text-right">
                                                    <h3>{{ cloudflareMeta.total_count }} total items</h3>
                                                </CCol>
                                            </CRow>
                                            <CDataTable
                                                :items="cloudflareBlacklist"
                                                :fields="['target', 'ip', 'comment', 'mode', 'created_on']"
                                                :items-per-page="cloudflareQuery.page_per_count"
                                                :pagination="false"
                                                :items-per-page-select="{values: [10,25,100,500]}"
                                                :hover="true"
                                                @pagination-change="(e) => handleCloudflarePerPageChange(e, 1)"
                                            >
                                            </CDataTable>
                                            <CPagination
                                                v-if="cloudflareMeta != null"
                                                :active-page="cloudflareMeta.page"
                                                :pages="cloudflareMeta.total_pages"
                                                @update:activePage="(newPage) => handleCloudflarePerPageChange(cloudflareQuery.page_per_count, newPage)"
                                            />
                                        </CCardBody>
                                    </CCard>
                                </CCol>
                            </CRow>
                        </CCardBody>
                    </CCard>
                </CCol>
            </CRow>
        </CContainer>

        <Toaster ref="toaster" />

        <!-- Confirm Modal for Executing Job -->
        <ConfirmModal
            v-if="confirmOptions"
            :isOpen="confirmOptions!=null"
            :options="confirmOptions"
            @confirm="confirmOptions.onConfirm(confirmOptions.job_name,confirmOptions.job_reason)"
            @cancel="closeConfirmModal"
        />
    </div>
</template>

<script>
import ConfirmModal from '@/components/ConfirmModal.vue'
import Toaster from '@/components/Toaster.vue'
import { CDataTable, CInput } from '@coreui/vue';
export default {
    name: "IPBlacklist",
    components: {
        ConfirmModal,
        Toaster
    },
    data (){
        return {
            blacklist: [],
            blacklistMeta: {},

            cloudflareBlacklist: [],
            cloudflareQuery:{
                page: 1,
                page_per_count: 25
            },
            cloudflareMeta: {},

            baninfo:{ "ip": "", "reason": ""},

            confirmOptions:null,
        };
    },
    computed: {
        blacklistParsed(){
            if (!this.blacklist) return [];
            return this.blacklist.map(item => ({
                ...item,
                _classes: (item.Deleted_at || item.DeletedAt) ? 'table-danger' : 'table-success'
            }));
        }
    },
    create(){
        
    },
    mounted (){
        this.getIPBlacklist();
        this.getCloudflareIPBlacklist();
    },
    methods: {

        async getIPBlacklist(){
            let urlstr = `${this.$gstate.defaultUrl}/serviceban/admin/ip-blacklist`;
            try{
                const resp = await this.axios.get(urlstr, this.$auth.apiOptions);               
                // console.log("resp:", resp);
                this.blacklist = resp.data.data.blacklist;
                // console.log("list:", this.blacklist);
                this.$refs.toaster.toast("Got IP Blacklist data!", "light");
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }            
        },
        async createIPBlacklist(ip,reason){
            let urlstr = `${this.$gstate.defaultUrl}/serviceban/admin/ip-blacklist`;
            try{
                const body = { "ip": ip, "reason": reason};
                const resp = await this.axios.post(urlstr, body, this.$auth.apiJsonOptions);               
                // console.log("resp:", resp);
                // console.log("list:", this.blacklist);
                // Safely extract timestamp or fallback to current local time
                const timestamp = resp?.data?.data?.ts || new Date().toLocaleString();
                this.$refs.toaster.toast(`Create/Update IP Ban successfully! ${ip} (${reason}) [${timestamp}]`, "success");
                await this.getIPBlacklist();
                this.closeConfirmModal();
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }            
        },
        async deleteIPBlacklist(ip){
            let urlstr = `${this.$gstate.defaultUrl}/serviceban/admin/ip-blacklist`;
            try{
                // const body = { "ip": ip };
                const config = {
                    ...this.$auth.apiJsonOptions,
                    data: { ip: ip }
                };
                const resp = await this.axios.delete(urlstr, config);               
                // console.log("resp:", resp);
                // Safely extract timestamp or fallback to current local time
                const timestamp = resp?.data?.data?.ts || new Date().toLocaleString();
                this.$refs.toaster.toast(`Delete IP Ban successfully! ${ip} [${timestamp}]`, "success");
                await this.getIPBlacklist();
                this.closeConfirmModal();
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
        },
        async recoverIPBan(ip){
            let urlstr = `${this.$gstate.defaultUrl}/serviceban/admin/ip-blacklist`;
            try{
                const body = {"ip": ip};
                // console.log(body);
                const resp = await this.axios.put(urlstr, body, this.$auth.apiJsonOptions);
                const timestamp = resp?.data?.data?.ts || new Date().toLocaleString();
                this.$refs.toaster.toast(`Recover IP Ban successfully! ${ip} [${timestamp}]`, "success");
                await this.getIPBlacklist();
                this.closeConfirmModal();
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
        },

        async getCloudflareIPBlacklist(){
            let urlstr = `${this.$gstate.defaultUrl}/serviceban/admin/cloudflare-ip-blacklist`;
            const config = {
                ...this.$auth.apiOptions,
                params: {
                    page: this.cloudflareQuery.page,
                    page_per_count: this.cloudflareQuery.page_per_count
                }
            };
            try{
                const resp = await this.axios.get(urlstr, config);               
                // console.log("resp:", resp);
                this.cloudflareBlacklist = resp.data.data.blacklist;
                this.cloudflareMeta = resp.data.data.meta;
                // console.log("list:", this.cloudflareBlacklist);
                // console.log("meta:", this.cloudflareMeta);
                this.$refs.toaster.toast("Got Cloudflare IP Blacklist data!", "light");
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }   
        },
        handleCloudflarePerPageChange(newPerPage, page) {
            // console.log("handleCloudflarePerPageChange", newPerPage, page);
            if (this.cloudflareQuery.page_per_count != newPerPage){
                // console.log("per page change", newPerPage);
                this.cloudflareQuery.page_per_count = newPerPage;
                this.cloudflareQuery.page = 1;
            }else if (this.cloudflareQuery.page != page){
                // console.log("page change", page);
                this.cloudflareQuery.page = page;
            }
            this.getCloudflareIPBlacklist();
        },

		closeConfirmModal(){
			this.confirmOptions = null;
		},
		openExecuteConfirmModal(itm){
            // console.log(itm);

            let toSoftDel = true;
            if(itm.DeletedAt!=null){
                toSoftDel = false;
            }

			this.confirmOptions = {
				title: toSoftDel ?'Confirm Soft Delete' : 'Confirm Recover IP Ban' ,
				message: toSoftDel ?'Are you sure you want to soft delete this IP Ban:' + itm.IP : 'Are you sure you want to recover this IP Ban:' + itm.IP,
				confirmText: 'Execute',
				confirmColor: 'danger',
				cancelText: 'Cancel',
				onConfirm: toSoftDel ? (jbNm, _) => this.deleteIPBlacklist(jbNm) : (jbNm, _) => this.recoverIPBan(jbNm),
				job_name: itm.IP
			};
		},
        openCreateConfirmModal(){
            this.confirmOptions = {
				title:'Confirm Create/Update IP Ban',
				message:'Are you sure you want to create/update this IP Ban: ' + this.baninfo.ip + ' (' + this.baninfo.reason + ')' ,
				confirmText: 'Create/Update',
				confirmColor: 'danger',
				cancelText: 'Cancel',
				onConfirm: (jbNm, jbRsn) => this.createIPBlacklist(jbNm, jbRsn),
				job_name: this.baninfo.ip,
                job_reason: this.baninfo.reason
			};
        }


    }
}
</script>