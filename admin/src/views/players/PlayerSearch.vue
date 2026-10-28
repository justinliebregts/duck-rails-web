<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1>Players</h1>
            </CHeader>
            <CRow>
                <CCol sm="3">
                    <CInput
                    label="Search user by name"
                    placeholder="Name"
                    :value.sync="playerSearchName"
                    @keyup.enter="searchPlayerByName()"
                    />
                    <CInput
                    label="Search user by ID"
                    placeholder="ID"
                    :value.sync="playerSearchId"
                    @keyup.enter="searchPlayerById()"
                    />  
                    <CButton
                        :color= "isSearchingPlayers ? 'dark' : 'info'"
                        @click="playerSearchId != '' ? searchPlayerById() : searchPlayerByName()"
                        class=""
                    >
                        Search
                    </CButton>
                </CCol>
                <CCol sm="9" class="text-right">
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=3724467550970486'">Lebolay</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=9502428739782353'">lebo_wtjc4k</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=3655906067778852'">Duckrice</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=3552842248147875'">Ken</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=4480519545359046'">testergr_j55cvk</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=2358325487625010'">Kota</CButton>
                    <CButton class="ml-3" color="secondary" :href="$gstate.getAdminUrl() + '/player/details?id=125673113972659'">bebo_6sd4q2</CButton>
                </CCol>
                
            </CRow>
			
            <CRow>
                <CDataTable
                    v-if="playersData != null"
                    :items="playersData"
                    :fields="fields"
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
                    :add-table-classes="{}"
                    :noItemsView="{ noResults: 'No filtering results available.', noItems: 'No players found.' }"
                />
            </CRow>

			<CRow v-if="onlinePlayers != null && playersData == null" >
				<CCol>
				  <h3 class="mt-5">Online Players</h3>
					<CDataTable
						:items="onlinePlayers"
						:fields="onlinePlayerfields"
						sorter
						column-filter
						hover
						:responsive="false"
						:items-per-page="10"
						:pagination="true"
						:items-per-page-select="true"
						v-on:row-clicked="rowClick"
					/>
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
import CoreUIIcons from '../icons/CoreUIIcons.vue';

export default {
    components: { CoreUIIcons },

    // chrome.exe --user-data-dir="C://Chrome dev session" --disable-web-security

    name: "Players",
    data (){
        return {
            playersData: null,
            fields: ['id', 'name', 'platform', 'ip', {'key':'is_converted', 'label':'Is Converted'}, 'created', 'updated', 'copy'],
            toastMsg: [],
            toastColor: [],
            fixedToasts: 0,
            playerSearchName: '',
            playerSearchId: '',
            clickCount: 0,
            timer: null,
            isSearchingPlayers: false,
			onlinePlayerfields: ['id'],
			onlinePlayers: []//[{"id":"123"},{"id":"456"}]
        };
    },
	mounted (){
        this.fetchOnlinePlayer();

        if(this.$route.query.searchName != undefined){
            this.playerSearchName = this.$route.query.searchName;
            this.searchPlayer(true, this.playerSearchName);
        }

    },
    methods: {
        searchPlayer: function (isSearchByName, searchValue) {
            this.isSearchingPlayers = true;
            let routeUrl = isSearchByName ? '/player/admin/search-by-name/' : '/player/admin/search-by-id/';
            let urlstr = this.$gstate.getBaseUrl() + routeUrl + searchValue;
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {

                if(response.data.result != undefined){
                    this.toast("Error: " + response.data.messages[0], "light");
                    this.playersData = null;
                    return;
                }
                this.playersData = response.data.data;
                
                for (var i in this.playersData) {
                    this.playersData[i].copy = "⧉";
                }

                if(this.playersData.length == 0){
                    this.toast("No players found :(", "light");
                } else {
                    this.toast("Got players!", "light");
                }
                if(this.playersData.length == 1) {
                    this.rowClick(this.playersData[0], null, null, null);
                }
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //console.log("final")
                this.isSearchingPlayers = false;
            })
            
        },
        searchPlayerById: function () {
            if(this.playerSearchId == ''){
                this.toast("ID required.", "danger");
                return;
            }
            this.searchPlayer(false, this.playerSearchId);           
        },
        searchPlayerByName: function (){
            
            if(this.playerSearchName == ''){
                this.toast("Name required.", "danger");
                return;
            }
            
            this.$router.push({path:"/player/search?searchName=" + this.playerSearchName});

            this.searchPlayer(true, this.playerSearchName);           
        },
		fetchOnlinePlayer: function () {
            let urlstr = this.$gstate.getBaseUrl() + "/player/admin/online";
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {

                if(response.data.result != undefined){
                    this.toast("Error: " + response.data.messages[0], "light");
                    this.playersData = null;
                    return;
                }
                //this.onlinePlayers = response.data.data;
				for (var i in response.data.data) {
					//console.log(i);
					console.log(response.data.data[i]);
					this.onlinePlayers.push({"id": response.data.data[i]}); 
					//console.log(this.onlinePlayers);
				}
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //console.log("final")
                //this.isSearchingPlayers = false;
            })          
        },
        rowClick: function(item, index, columnName, event){

            if(columnName == "copy"){
                this.copyToClipboard(item.id);
                return;
            }
            
            this.clickCount++;
            // if double clicking, it opens in new tab
            if (this.clickCount === 1) {
                this.timer = setTimeout(() => {
                    //console.log('Single click detected');
                    this.$router.push({path:"/player/details?id=" + item.id});
                    this.clickCount = 0;
                }, 250);
            } else if (this.clickCount === 2) {
                clearTimeout(this.timer);
                //console.log('Double click detected');
                let id = item.id;
                const url = this.$router.resolve({ path: `/player/details`, query: { id } }).href;

                // Open the URL in a new tab
                window.open(url, '_blank');
                this.clickCount = 0;
            }

        },
        toast: function(msg, color = "primary"){
            
            this.fixedToasts++;
            this.toastMsg[this.fixedToasts] = msg;
            this.toastColor[this.fixedToasts] = color;
            
        },
        async copyToClipboard(textToCopy) {
            try {
                await navigator.clipboard.writeText(textToCopy);
                this.toast("ID saved to clipboard", "success");
            } catch (err) {
                console.error('Failed to copy:', err);
            }
        },
    }

};
</script>