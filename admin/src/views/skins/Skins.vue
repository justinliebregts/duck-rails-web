<template>
    <div>
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1>Skins</h1>
            </CHeader>
            <CRow>
                <CCol sm="6">
                    <CButton
                    color="info"
                    v-on:click="createSkinButtonClick()"
                    class="mb-3 mr-3"
                    >
                        Create New Skin
                        
                    </CButton>
                    <CButton
                    color="secondary"
                    v-on:click="getSkins()"
                    class="mb-3"
                    >
                        Refresh
                    </CButton>
                </CCol>
                <CCol v-if="skinsData" sm="6" class="text-right">
                    <h3>{{ skinsData.length }} total skin styles</h3>
                </CCol>
            </CRow>
            

            <CModal v-if="focusedSkin != null"
                title="Create New Skin"
                color="info"
                :show.sync="createSkinModal"
                :centered="true"
                size="sm"
            >
                
                <CRow>
                    <CCol sm="12">
                        <CInput
                        label="Id"
                        :value.sync="focusedSkin.id"
                        description="Must match the C# skin style enum. Leave blank to use the next id."
                        />
                        <CInput
                        label="Name"
                        :value.sync="focusedSkin.name"
                        />
                    </CCol>
                </CRow>
                
                <template #footer>
                    <CButton @click="createSkinModal = false" color="light">Cancel</CButton>
                    <CButton @click="createSkin()" color="success">Create Skin</CButton>
                </template>
            </CModal>


            <CModal v-if="focusedSkin != null"
                :title="focusedSkin.name"
                color="secondary"
                :show.sync="skinModal"
                :centered="true"
                size="xl"
            >
                
                <CRow>
                    <CCol sm="6">
                        <CInput
                        label="Id"
                        :value.sync="focusedSkin.id"
                        :readonly="true"
                        description="ID must match the ID of the skin_styles enum in C#"
                        />
                    </CCol>
                    <CCol sm="6">
                        <CInput
                        label="Name"
                        :value.sync="focusedSkin.name"
                        />
                    </CCol>
                </CRow>
                
                <template #footer>
                    <CButton @click="skinModal = false" color="light">Cancel</CButton>
                    <CButton @click="saveSkin()" color="success">Save Skin</CButton>
                </template>
            </CModal>

            <CDataTable
                    :items="skinsData"
                    :fields="fields"
                    sorter
                    column-filter
                    hover
                    :responsive="false"
                    :items-per-page="50"
                    :pagination="true"
                    :items-per-page-select="true"
                    v-on:row-clicked="rowClick"
                    :loading="skinsData == null"
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

export default {

    // chrome.exe --user-data-dir="C://Chrome dev session" --disable-web-security

    name: "Skins",
    //components: { ColorTheme },
    data (){
        return {
            createSkinModal: false,
            skinModal: false,
            focusedSkin: null,
            skinsData: null,
            fields: ['id', 'name'],
            toastMsg: [],
            toastColor: [],
            fixedToasts: 0
        };
    },
    create(){
        
    },
    mounted (){
        this.getSkins();
    },
    methods: {
        toast: function(msg, color = "primary"){
            
            this.fixedToasts++;
            this.toastMsg[this.fixedToasts] = msg;
            this.toastColor[this.fixedToasts] = color;
            
        },
        saveSkin: function (){
            
            this.skinModal = false;

            let urlstr = this.$gstate.getBaseUrl() + '/skin/admin/update/' + this.focusedSkin.id;

            var data = new FormData();
            data.append('name', this.focusedSkin.name);

            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(response => {

                this.toast(this.focusedSkin.name + ": " + response.data.messages[0], "success");
                this.focusedSkin = null;
                
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //
            })


        },
        createSkinButtonClick: function(){
            this.createSkinModal = true;
            this.focusedSkin = {};
        },
        createSkin: function (){
            
            this.createSkinModal = false;

            let urlstr = this.$gstate.getBaseUrl() + '/skin/admin/create';

            var data = new FormData();
            if (this.focusedSkin.id) {
                data.append('id', this.focusedSkin.id);
            }
            data.append('name', this.focusedSkin.name);

            this.axios.post(urlstr, data, this.$auth.apiOptions)
            .then(response => {
                this.toast(response.data.data + " " + this.focusedSkin.name + ": " + response.data.messages[0], "success");

                setTimeout(()=> {
                    this.toast("Remember to add this to the C# enum!", "info");
                }, 500)
                
                let newSkin = {"id" : response.data.data, "name" : this.focusedSkin.name}
                this.skinsData.push(newSkin);
                this.focusedSkin = null;
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //
            })


        },
        rowClick: function(item, index, columnName, event){
            this.skinModal = true;
            this.focusedSkin = item;
        },
        getSkins: function (){

            let urlstr = this.$gstate.getBaseUrl() + '/skin/admin/all';
            this.axios.get(urlstr, this.$auth.apiOptions)
            .then(response => {
                this.skinsData = response.data.data;   
                
                for (let i = 0; i < this.skinsData.length; i++) {
                    this.skinsData[i].id = parseInt(this.skinsData[i].id);
                }
            })
            .catch(error => {
                console.log(error)
                if(error.response != undefined && error.response.data != undefined) {this.toast(error.response.data.messages[0], "danger") }
                this.toast(error, "danger");
            })
            .finally(() => {
                //console.log("final")
            })
        }
    }

};
</script>
