<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1 v-if="playerData">{{ playerData.name }} <small style="font-size:0.85rem">{{ playerData.id }}</small></h1>
            </CHeader>

            <DuckCoinsCard :player-id="playerId" @toast="toast" />
            <CCard v-if="playerData">
                <CCardBody>
                    <CRow>
                        <CCol lg="2">
                            Platform: <small><b>{{ playerData.platform }}</b></small><br />
                            Device: <small><b>{{ playerData.device }}</b></small><br />
                            Version: <small><b>{{ playerData.version }}</b></small><br />
                            IP: <small><b>{{ playerData.ip }}</b></small><br />
                            Device UUID: <small><b>{{ playerData.device_uuid }}</b></small><br />
                        </CCol>
                        <CCol lg="2">
                            Created: <small><b>{{ playerData.created }}</b></small><br />
                            Last Played: <small><b>{{ playerData.updated }}</b></small><br />
                        </CCol>
                        <CCol lg="3">
                        </CCol>
                    </CRow>
                </CCardBody>
            </CCard>

            <CRow>
                <CCol sm="12">
                    <CButton
                        v-if="playerData"
                        color="success"
                        v-on:click="openGiveItem"
                        class="mr-3 mb-3"
                    >
                        Give Item
                    </CButton>
                    <CButton
                        color="danger"
                        v-on:click="deleteAllTransactionsConfirm = true"
                        class="mr-3 mb-3"
                    >
                        Delete all transactions
                    </CButton>
                    <CButton
                        color="info"
                        v-on:click="setServiceBanModal(true)"
                        class="mr-3 mb-3"
                    >
                        Show ServiceBan
                    </CButton>
                    <CButton
                      color="info"
                      v-show="playerData && playerData.platform === 'OCULUS'"
                      v-on:click="goToAttestationDetail"
                      class="mr-3 mb-3"
                    >
                        GoTo Attestation Page
                    </CButton>
                </CCol>
            </CRow>

            <CModal
                title="Give Item"
                color="success"
                :centered="true"
                size="sm"
                :show.sync="giveItemModal"
            >
                <CInput
                    label="Item ID"
                    placeholder="items.id"
                    :value.sync="giveItemId"
                />
                <template #footer>
                    <CButton @click="giveItemModal = false" color="light">Cancel</CButton>
                    <CButton @click="giveItem()" color="success">Give</CButton>
                </template>
            </CModal>

            <CModal 
                title="Delete All Transactions for real?"
                color="danger"
                :centered="true"
                size="sm"
                :show.sync="deleteAllTransactionsConfirm"
                
            >
                <div v-if="playerData">Really delete all <b>{{ playerName }}'s</b> transactions?</div>
                
                <template #footer>
                    <CButton @click="deleteAllTransactionsConfirm = false" color="light">Cancel</CButton>
                    <CButton @click="deleteAllTransactions()" color="danger">Delete all {{ playerName }}'s transactions</CButton>
                </template>
            </CModal>

			<CModal 
                title="ServiceBan List"
                color="info"
                :centered="true"
                size="xl"
                :show.sync="showServiceBanModal"
            >

                <CRow>
                    <CCol>
						<CDataTable
							:items="sbItems_computed"
							:fields="sbcolumns"
							v-on:row-clicked="sbRowClick"
							sorter
							hover
							:responsive="true"
							:pagination="false"
							:items-per-page-select="false"
						/>
                    </CCol>
                </CRow>

          <template #header>
            <CButton v-if="!showExpiredBans" @click="setShowExpiredBans(true)" color="light">Show Expired Bans</CButton>
            <CButton v-if="showExpiredBans" @click="setShowExpiredBans(false)" color="danger">Hide Expired Bans</CButton>
          </template>
                <template #footer>
                    
					<CButton @click="setServiceBanCreation(true)" color="light">Show Creation Panel</CButton>
					<CButton @click="setServiceBanModal(false)" color="danger">Cancel</CButton>
                </template>

				<CCollapse :show="showServiceBanCreation">
				<br/>
				<CRow>
				<CCol sm="4">
					<CSelect
					  label="TargetType"					  
					  v-model="serviceBanTarget"
					  :options="['platform','device']"
					/>
				</CCol>
				<CCol sm="4">
					<CSelect
					  label="BanType"
					  v-model="serviceBanType"
					  :options="['game','mic']"
					/>
				</CCol>
				<CCol>
                        <CInput
                            label="Ban duration in days"
                            placeholder="days"
                            :value.sync="banDays"
                            />
                    </CCol>
				</CRow>
				<CRow>
				<CCol sm="12">
					<CTextarea
					  label="Reason"
					  :value.sync="serviceBanReason"
					  placeholder="{}"
					/>
				  </CCol>
				</CRow>
				<CRow>
				  <CButton @click="setServiceBanCreation(false)" variant="ghost" color="dark">Hide</CButton>
				  <CButton @click="resetServiceBan()" variant="ghost" color="dark">Reset</CButton>
				  <CButton @click="createServiceBan()" color="danger">Create {{this.serviceBanTarget}} - BanType:{{this.serviceBanType}} for {{banDays}} days</CButton>
				</CRow>
				</CCollapse>
            </CModal>
				
			<CModal v-if="sbToDelete"
				backdrop="static"
                title="Delete ServiceBan?"
                :color="modalColor"
                :centered="true"
                size="sm"
                :show.sync="deleteServiceBanModal"
                
            >
                Delete ServiceBan <br/> {{this.sbToDelete.id}} <br/> {{this.sbToDelete.target_type}} {{this.sbToDelete.target_id}} <br/> BanType:{{this.sbToDelete.ban_type}}
                <template #footer>
                    <CButton @click="deleteServiceBanModal = false" color="light">Cancel</CButton>
                    <CButton @click="deleteServiceBan()" :color="modalColor">Delete</CButton>
                </template>
            </CModal>
            
            <CModal v-if="txnToDelete"
                :title='modalAction+"?"'
                :color="modalColor"
                :centered="true"
                size="lg"
                :show.sync="deleteTxnModal"
                
            >
                <div v-html="deleteTxnMessage"></div>
                
                <template #footer>
                    <CButton @click="deleteTxnModal = false" color="light">Cancel</CButton>
                    <CButton @click="deleteTransaction()" :color="modalColor">{{ modalAction }}</CButton>
                </template>
            </CModal>

            <CRow>
                <CCol sm="12">
                    <CDataTable
                        :items="transactions"
                        :fields="transactionFields"
                        sorter
                        column-filter
                        hover
                        :clickable-rows="true"
                        :responsive="false"
                        :items-per-page="20"
                        :pagination="true"
                        :items-per-page-select="true"
                        v-on:row-clicked="rowClick"
                        :loading="isSearchingPlayers"
                        :small="true"
                        :cleaner="true"
                        :noItemsView="{ noResults: 'No filtering results available.', noItems: 'No transactions found.' }"
                    >
                        <template #transaction_is_deleted="{item}">
                            <td>
                                <CBadge v-if="item.transaction_is_deleted" color="danger">Deleted</CBadge>
                            </td>
                        </template>
                    </CDataTable>
                </CCol>
            </CRow>

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

            
        </CContainer>
    </div>
</template>

<script>
import DuckCoinsCard from "./DuckCoinsCard.vue";
export default {
  components: { DuckCoinsCard },
  name: "PlayerDetails",
  data (){
    return {
      playerData: null,
      giveItemModal: false,
      giveItemId: "",
      transactions: [],
      transactionFields: [
        { key: 'transaction_id', label: 'Txn ID' },
        { key: 'item_id', label: 'Item ID' },
        { key: 'item_name', label: 'Item' },
        { key: 'skin_name', label: 'Skin' },
        { key: 'platform', label: 'Platform' },
        { key: 'platform_txn_id', label: 'Receipt' },
        { key: 'transaction_is_deleted', label: 'Deleted' },
        { key: 'transaction_date', label: 'Date' },
      ],
      toastMsg: [],
      toastColor: [],
      fixedToasts: 0,
      isSearchingPlayers: false,
      playerId: null,
      deleteTxnModal: true,
      txnToDelete: null,
      actionToDelete: true,
      deleteAllTransactionsConfirm: false,
      showExpiredBans: false,
      showServiceBanModal: false,
      banDays: 0,
			sbcolumns:[
          {
            key: 'id',
            label: '#',
            _props: { scope: 'col' },
          },
          {
            key: 'target_type',
            _props: { scope: 'col' },
          },
          {
            key: 'target_id',
            _props: { scope: 'col' },
          },
          {
            key: 'ban_type',
            _props: { scope: 'col' },
          },
          {
            key: 'reason',
            _props: { scope: 'col' },
          },
          {
            key: 'start_datetime',
            _props: { scope: 'col' },
          },
          {
            key: 'end_datetime',
            _props: { scope: 'col' },
          },
      ],
			sbItems:[],
			deleteServiceBanModal: false,
			sbToDelete: null,
			showServiceBanCreation: false,
			serviceBanTarget: 'device',
			serviceBanType: 'mic',
			serviceBanReason:'{}',
    };
    },
    computed: {
        sbItems_computed() {
            if(this.showExpiredBans) {
              return this.sbItems;
            } else {
              return this.sbItems.filter(item => item.exp === false);
            }
        },

        deleteTxnMessage(){

            if(!this.txnToDelete){
                return "No transaction to edit...";
            }

            const name = this.txnToDelete.item_name || this.txnToDelete.item_id || "";
            if(this.actionToDelete){
                return "Really delete <b>" + this.txnToDelete.transaction_id + " - " + name + "</b>?";
            }
            return "Really restore <b>" + this.txnToDelete.transaction_id + " - " + name + "</b>?";
        },
        modalColor(){
            if(this.actionToDelete){
                return "danger";
            } else {
                return "info";
            }
        },
        modalAction(){
            if(this.actionToDelete){
                return "Delete Transaction";
            } else {
                return "Restore Transaction";
            }
        },
        playerName(){
            if(this.playerData == undefined){
                return "";
            } else {
                return this.playerData.name;
            }
        },
    },
    mounted (){
    this.playerId = this.decodeHTMLEntities(this.$route.query.id);
    if(this.playerId == null || this.playerId == undefined){
      this.$router.push({path:"/player/search"});
    }

    this.getTransactionData();
    this.getPlayerData();
		this.getServiceBanData();
    },
    methods: {
        openGiveItem: function () {
            this.giveItemId = "";
            this.giveItemModal = true;
        },
        giveItem: function () {
            if (!this.playerData || !this.giveItemId) {
                this.toast("Item ID is required.", "danger");
                return;
            }
            const data = new FormData();
            data.append("player_id", this.playerData.id);
            data.append("item_id", this.giveItemId);
            this.axios.post(this.$gstate.getBaseUrl() + "/inventory/admin/give-item", data, this.$auth.apiOptions)
                .then(response => {
                    this.giveItemModal = false;
                    this.getTransactionData();
                    const msg = response.data.messages && response.data.messages[0];
                    this.toast(msg || "Item given", "success");
                })
                .catch(error => {
                    console.log(error);
                    if (error.response != undefined && error.response.data != undefined) {
                        this.toast(error.response.data.messages[0], "danger");
                        return;
                    }
                    this.toast(error, "danger");
                });
        },

		setServiceBanModal: function(value){
			this.showServiceBanModal = value;
        },
		setServiceBanCreation: function(value){
			this.showServiceBanCreation=value;
		},
        decodeHTMLEntities: function(text) {
            var entities = [
                ['amp', '&'],
                ['apos', '\''],
                ['#x27', '\''],
                ['#x2F', '/'],
                ['#39', '\''],
                ['#47', '/'],
                ['lt', '<'],
                ['gt', '>'],
                ['nbsp', ' '],
                ['quot', '"'],
                ['%7C', '|']
                
            ];

            for (var i = 0, max = entities.length; i < max; ++i) 
                text = text.replace(new RegExp('&'+entities[i][0]+';', 'g'), entities[i][1]);

            return text;
        },
      setShowExpiredBans: function (value) {
        this.showExpiredBans = value;
      },

        getPlayerData: function (){

            let urlstr = this.$gstate.getBaseUrl() + '/player/admin/get/' + this.playerId;
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {

                if(response.data.message != undefined && response.data.result == false){
                    this.toast("Error: " + response.data.messages[0], "danger");
                    return;
                }

                this.playerData = response.data.data;
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
        },

		getServiceBanData: function (){

            let urlstr = this.$gstate.getBaseUrl() + '/serviceban/admin/get/' + this.playerId;
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {

                if(response.data.message != undefined && response.data.result == false){
                    this.toast("Error: " + response.data.messages[0], "danger");
					this.sbItems=[];
                    return;
                }

                this.sbItems = response.data.data;
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
        },
        getTransactionData: function (){
            const urlstr = this.$gstate.getBaseUrl() + '/admin/transactions/' + this.playerId;
            this.isSearchingPlayers = true;
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {
                this.transactions = response.data.data || [];
            })
            .catch(error => {
                console.log(error);
                const data = error && error.response && error.response.data;
                this.toast((data && (data.error || (data.messages && data.messages[0]))) || 'Failed to load transactions', "danger");
            })
            .finally(() => {
                this.isSearchingPlayers = false;
            });
        },
        toast: function(msg, color = "primary"){
            
            this.fixedToasts++;
            this.toastMsg[this.fixedToasts] = msg;
            this.toastColor[this.fixedToasts] = color;
            
        },
        rowClick: function(item){

            if(item.transaction_is_deleted){
                this.actionToDelete = false;
            } else {
                this.actionToDelete = true;
            }

            this.deleteTxnModal = true;
            this.txnToDelete = item;
        },
		sbRowClick: function(serviceBan){
            this.deleteServiceBanModal = true;
            this.sbToDelete = serviceBan;
        },
		deleteServiceBan: function (){

            if(!this.sbToDelete){
                this.toast("ServiceBan not found.", "danger");
                return;
            }

            let urlstr = this.$gstate.getBaseUrl() + '/serviceban/admin/end/' + this.sbToDelete.id;

            var data = new FormData();
            data.append('is_deleted', this.actionToDelete ? "1" : "0");

            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(response => {
				this.getServiceBanData();
				this.toast(response.data.message, "success");
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                this.sbToDelete = null;
				this.deleteServiceBanModal = false;
            })
        },
		resetServiceBan: function(){
			this.serviceBanTarget= 'device';
			this.serviceBanType= 'mic';
			this.serviceBanReason= '{}';
		},
		createServiceBan: function(){
            let urlstr = this.$gstate.getBaseUrl() + '/serviceban/admin/create';
            
            var data = new FormData();
			data.append('target_type', this.serviceBanTarget);
			switch (this.serviceBanTarget) {
				case 'platform':
				data.append('target_id', this.playerData.id);
				break;
				case 'device':
				data.append('target_id', this.playerData.device_uuid);
				break;
			}
            data.append('ban_type', this.serviceBanType);
            data.append('duration', this.banDays);
			data.append('reason', this.serviceBanReason);

            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(response => {

                this.setServiceBanCreation(false);
				this.getServiceBanData();
                this.toast(response.data.messages[0], "success");
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) { this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
		},
        deleteTransaction: function (){

            if(!this.txnToDelete){
                this.toast("Transaction not found.", "danger");
                return;
            }

            let urlstr = this.$gstate.getBaseUrl() + '/inventory/admin/delete/' + this.txnToDelete.transaction_id;

            var data = new FormData();
            data.append('is_deleted', this.actionToDelete ? "1" : "0");

            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(response => {

                this.txnToDelete = null;
                    this.deleteTxnModal = false;
                    this.getTransactionData();
                    this.toast(response.data.messages[0], "success");
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
        },
      goToAttestationDetail() {
        if (this.playerData.platform !== "OCULUS") return;
        this.$router.push({
          name: 'PlatformId-uniqueIdDetail',
          params: { platform_id: this.playerData.id }
        });
      },
        deleteAllTransactions: function (){

            let urlstr = this.$gstate.getBaseUrl() + '/inventory/admin/delete-all';            
            
            var data = new FormData();
            data.append('player_id', this.playerData.id);
            
            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(() => {

                this.toast("All transactions deleted!", "success");
                this.deleteAllTransactionsConfirm = false;
                this.getTransactionData();

            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
        },
    }

};
</script>
