<template>
    <div>
        
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1>Active Game Bans</h1>
            </CHeader>
            <CRow>
                <CCol sm="3">
                </CCol>
                <CCol sm="6">
                    <CRow>
                        <CCol sm="9">
                            <CInput type="text" placeholder="Search by Device ID, Player ID, Player Name, Player IP..." v-model="banQuery.query.q" @keyup.enter="onSearchSubmit"></CInput>
                        </CCol>
                        <CCol sm="3">
                            <CButton color="secondary" @click="resetFilters" v-if="banQuery.query.q.length > 0 || banQuery.query.page != 1">
                                Reset
                            </CButton>
                        </CCol>
                    </CRow>
                </CCol>
                <CCol sm="3" class="text-right" v-if="bansMeta != null">
                    Total Counts: {{ bansMeta.total_count }} | Total Pages: {{ bansMeta.total_pages }}
                </CCol>
            </CRow>
            <CDataTable
                :items="bansData"
                :fields="fields"
                hover
                :responsive="false"

                :items-per-page-select="{values: [10, 50, 100, 200, 500, 5000]}"
                :items-per-page="banQuery.query.page_per_count"
                @pagination-change="onPageSizeChange"

                :pagination="false"
                :loading="bansData == null"
                :small="true"
                :cleaner="true"
                v-on:row-clicked="rowClick"
                :add-table-classes="{}"
                :noItemsView="{ noResults: 'No filtering results available.', noItems: 'No active game bans.' }"
            >
                <template #seconds_remaining="{item}">
                    <td>{{ item.time_remaining }}</td>
                </template>

                <template #target_id="{ item }">
                    <td>
                        <span v-html="highlightText(item.target_id)"></span>
                    </td>
                </template>
                <template #player_name="{ item }">
                    <td>
                        <span v-html="highlightText(item.player_name)"></span>
                    </td>
                </template>
                <template #ip="{ item }">
                    <td>
                        <span v-html="highlightText(item.ip)"></span>
                    </td>
                </template>

            </CDataTable>

            <CPagination
                v-if="bansMeta != null"
                :active-page.sync="banQuery.query.page"
                :pages="bansMeta.total_pages"
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
        </CContainer>
    </div>
</template>

<script>

// import { ListQuery } from '@/models/ListQuery';

// class BanListQuery extends ListQuery {
//     constructor(limit) {
//         super(limit); // 初始化父類別構造函數
//         this.active_only = false;
//         this.ban_type = 'game';
//     }

//     reset() {
//         super.reset();
//         this.active_only = false;
//         this.ban_type = 'game';
//     }
// }
import { useBanListQuery } from './useBanListQuery.js';
import { getCurrentInstance } from '@vue/composition-api';

export default {
    // chrome.exe --user-data-dir="C://Chrome dev session" --disable-web-security

    name: "Bans",
    data (){

        return {
            bansData: null,
            bansMeta: null,
            fields: [
                'id',
                {'key':'target_type', 'label':'Target Type', _style:'min-width: 120px;'},
                {'key':'target_id', 'label':'Target ID'},
                {'key':'player_name', 'label':'Name'},
                {'key':'ban_type', 'label':'Type'},
                {'key':'start_datetime', 'label':'Ban Start'},
                {'key':'end_datetime', 'label':'Ban End'},
                {'key':'seconds_remaining', 'label':'Time Remaining'},
                {'key':'player_ip', 'label':'Player IP'},
                {'key':'device_uuid', 'label':'Device ID'},
            ],
            toastMsg: [],
            toastColor: [],
            fixedToasts: 0,

            // banQuery: new BanListQuery(100),
        };        
    },
    setup() {
        const { proxy } = getCurrentInstance();

        // 傳入安全的內部觸發函式，確保直接調用目前組件實體的 getBans
        const banQuery = useBanListQuery(() => {
            if (proxy && typeof proxy.getBans === 'function') {
                proxy.getBans();
            }
        }, 100, 'game');

        return {
            banQuery
        };
    },
    watch: {
        // 'banQuery.page': function (newPage, oldPage) {
        //     if (newPage == oldPage) 
        //         return;

        //     this.getBans();
        // },
        // 'banQuery.q': function (newVal, oldVal) {
        //     if (newVal === oldVal) 
        //         return;
            
        //     this.banQuery.handleSearchChange(this.getBans);
        // }
    },
    computed: {
        
    },
    create(){
        
    },
    mounted (){
        this.getBans();
    },
    methods: {
        onPageSizeChange(newSize) {
            this.banQuery.setPageSize(Number(newSize));
        },

        onSearchSubmit() {
            // Enter 鍵直接觸發提交與檢查頁碼
            this.banQuery.submitAndFetch();
        },

        resetFilters() {
            // 一鍵重置所有過濾條件與搜尋
            this.banQuery.resetAndFetch();
        },

        highlightText(text) {
            if (!text) return '';
            const keyword = this.banQuery.query?.q ? this.banQuery.query.q.trim() : '';
            if (!keyword) return String(text);

            const escapedKeyword = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
            const regex = new RegExp(`(${escapedKeyword})`, 'gi');
            return String(text).replace(regex, '<mark class="p-0 bg-warning text-dark">$1</mark>');
        },

        getBans: function (){

			let urlstr = this.$gstate.defaultUrl + '/serviceban/admin/all';

            //console.log("quer:", this.pageQuery);
            this.axios.get(urlstr, this.$auth.queryWithJsonOptions(this.banQuery.toParams()))
            .then(response => {
                this.bansData = response.data.data;
                this.bansMeta = response.data.meta;
                // console.log("bansMeta", this.bansMeta);
                this.toast("Got bans!", "light");
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //console.log("final")
            })
        },
        rowClick: function(item){
            if(item.target_type !== 'platform' || !item.target_id){
                this.toast("No player details for device bans.", "warning");
                return;
            }
            this.$router.push({path:"/player/details?id=" + item.target_id});
        },
        toast: function(msg, color = "primary"){
            
            this.fixedToasts++;
            this.toastMsg[this.fixedToasts] = msg;
            this.toastColor[this.fixedToasts] = color;
            
        },
    }

};
</script>
