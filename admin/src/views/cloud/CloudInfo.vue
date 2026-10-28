<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3 h1">
                Cloud Services Info
            </CHeader>
            <CRow>
                <CCol sm="12">
                    <CCard>
                        <CCardHeader class="h3"> Available Cron Jobs</CCardHeader>
                        <CCardBody>
                    <CDataTable
                        :items="availableCronJobs"
                        :fields="['name', 'schedule', 'suspend']"
                        :items-per-page="10"
                        :hover="true"
                        @row-clicked="openExecuteConfirmModal"
                    />
                    </CCardBody>
                    </CCard>
                </CCol>

                <CCol sm="12">
                    <CCard>
                        <CCardHeader class="h3"> Executed Jobs</CCardHeader>
                        <CCardBody>
                    <CDataTable
                        :items="executedJobs"
                        :fields="['name', 'status', 'start_time', 'end_time']"
                        :items-per-page="50"
                        :hover="true"
                        :footer="true"
                        @row-clicked="openShowLogConfirmModal"
                    />
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

            @confirm="confirmOptions.onConfirm(confirmOptions.job_name)"
            @cancel="closeConfirmModal"
        />
    </div>
</template>

<script>
import ConfirmModal from '@/components/ConfirmModal.vue'
import Toaster from '@/components/Toaster.vue'
import { CDataTable } from '@coreui/vue';
export default {
    name: "CloudInfo",
    components: {
        ConfirmModal,
        Toaster
    },
    data (){
        return {
            availableCronJobs: [],
            executedJobs: [],

            confirmOptions:null,

            loadedJobLog:null,
        };
    },
    computed: {
    },
    create(){
        
    },
    mounted (){
        this.getCronJobs();
        this.getExecutedJobs();
    },
    methods: {

        async getCronJobs(){
			const ns = this.$gstate.k8sNamespace || 'default';
            let urlstr = `${this.$gstate.defaultUrl}/info/admin/namespaces/${ns}/cronjobs`;
            try{
                const resp = await this.axios.get(urlstr, this.$auth.apiOptions);               
                //console.log("resp:", resp);
                this.availableCronJobs = resp.data.data.cronjobs;
                // console.log("availableCronJobs:", this.availableCronJobs);
                this.$refs.toaster.toast("Got cron jobs data!", "light");
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
            
        },
        async executedJob(cronjob_name){
            const ns = this.$gstate.k8sNamespace || 'default';
            const cj = cronjob_name || '';
            let urlstr = `${this.$gstate.defaultUrl}/info/admin/namespaces/${ns}/cronjobs/${cj}/trigger`;
            try{                
                //const resp = await this.axios.post(urlstr, null, this.$auth.apiOptions);               
                //console.log("resp:", resp);
                //this.executedJobs = resp.data.data.result;
                this.$refs.toaster.toast("Execute job:" + job_name, "light");
                this.getExecutedJobs();
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
            
        },
        async getExecutedJobs(){
			const ns = this.$gstate.k8sNamespace || 'default';
            let urlstr = `${this.$gstate.defaultUrl}/info/admin/namespaces/${ns}/jobs`;
            try{
                const resp = await this.axios.get(urlstr, this.$auth.apiOptions);               
                //console.log("resp:", resp);
                this.executedJobs = resp.data.data.jobs;
                //console.log("executedJobs:", this.executedJobs);
                this.$refs.toaster.toast("Got executed jobs data!", "light"); 
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
            
        },
        async loadJobLog(job_name){            
            const ns = this.$gstate.k8sNamespace || 'default';
            const job = job_name || '';
            let urlstr = `${this.$gstate.defaultUrl}/info/admin/namespaces/${ns}/jobs/${job}/logs`;
            // console.log("urlstr:", urlstr);
            try{
                const resp = await this.axios.get(urlstr, this.$auth.apiOptions);               
                //console.log("resp:", resp);
                this.loadedJobLog = resp.data.data.logs;
                this.$refs.toaster.toast("Execute job:" + job_name, "light");
            }catch(error){
                console.error(error);
                this.$refs.toaster.toast(error.message, "danger"); 
            }
            
        },

		closeConfirmModal(){
			this.confirmOptions = null;
		},
		openExecuteConfirmModal(cronjob){
			const cronjob_name = cronjob.name;
			this.confirmOptions = {
				title: 'Confirm Manual Trigger CronJob',
				message: 'Are you sure you want to manual trigger this Cronjob: ' + cronjob_name + '?',
				confirmText: 'Execute',
				confirmColor: 'danger',
				cancelText: 'Cancel',
				onConfirm: this.confirmExecuteJob,
				job_name: cronjob_name
			};
		},
        openShowLogConfirmModal(job){
            const job_name = job.name;
			this.confirmOptions = {
				title: 'Confirm Show Log',
				message: 'Are you sure you want to show log of this job?',
				confirmText: 'Show',
				cancelText: 'Cancel',
                onConfirm: this.confirmShowJobLog,
				job_name: job_name
			};
		},
		async confirmExecuteJob(jbName){
			await this.executedJob(jbName);
			this.closeConfirmModal();
		},
		async confirmShowJobLog(jbName){
			await this.loadJobLog(jbName);
			//this.closeConfirmModal();
            this.confirmOptions = {
				title: 'Log For Job ' + jbName ,
				message: 'Log: ' + this.loadedJobLog,
                confirmText: '',
                
				cancelText: 'Close',
				handleCancel: () => {this.closeConfirmModal(); this.loadedJobLog = null},
			};
		},

    }
}
</script>