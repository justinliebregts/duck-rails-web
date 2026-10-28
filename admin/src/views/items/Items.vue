<template>
    <div>
        <CContainer>
            <CHeader class="mb-3 p-3">
                <h1>Items</h1>
            </CHeader>
            <CRow>
                <CCol sm="6">
                    <CButton color="info" class="mb-3 mr-3" v-on:click="openCreate()">Create Item</CButton>
                    <CButton color="secondary" class="mb-3" v-on:click="load()">Refresh</CButton>
                </CCol>
                <CCol v-if="items" sm="6" class="text-right">
                    <h3>{{ items.length }} avatar and weapon skins</h3>
                </CCol>
            </CRow>

            <CDataTable
                :items="items"
                :fields="fields"
                sorter
                column-filter
                hover
                :clickable-rows="true"
                :responsive="false"
                :pagination="true"
                :items-per-page="25"
                v-on:row-clicked="openEdit"
                :noItemsView="{ noResults: 'No matching items.', noItems: 'No items yet.' }"
            />

            <CModal
                :title="editing ? 'Edit Item' : 'Create Item'"
                color="info"
                :show.sync="modal"
                :centered="true"
                size="lg"
            >
                <CInput label="Name" :value.sync="form.name" />
                <CSelect
                    v-if="!editing"
                    label="Type"
                    :value.sync="form.type"
                    :options="typeOptions"
                />
                <CSelect
                    v-if="form.type == 2"
                    label="Weapon"
                    :value.sync="form.weapon_type"
                    :options="weaponOptions"
                />
                <CSelect
                    label="Skin style"
                    :value.sync="form.skin_style_id"
                    :options="skinOptions"
                />
                <CInput label="Asset ID" :value.sync="form.asset_id" />
                <CSelect label="Default owned" :value.sync="form.is_default" :options="yesNo" />
                <CSelect label="Dev owned" :value.sync="form.dev_owned" :options="yesNo" />
                <CSelect label="Can own multiple" :value.sync="form.can_own_multiple" :options="yesNo" />
                <CSelect label="Ready" :value.sync="form.gr_ready" :options="yesNo" />
                <template #footer>
                    <CButton @click="modal = false" color="light">Cancel</CButton>
                    <CButton @click="save()" color="success">{{ editing ? "Save" : "Create" }}</CButton>
                </template>
            </CModal>

            <CToaster :autohide="2000">
                <template v-for="toast in fixedToasts">
                    <CToast :key="'toast' + toast" :show="true" :color="toastColor[toast]">
                        {{ toastMsg[toast] }}
                    </CToast>
                </template>
            </CToaster>
        </CContainer>
    </div>
</template>

<script>
function blankForm() {
    return {
        id: null,
        name: "",
        type: 3,
        weapon_type: "",
        skin_style_id: "",
        asset_id: "",
        is_default: 0,
        dev_owned: 0,
        can_own_multiple: 0,
        gr_ready: 1,
    };
}

export default {
    name: "Items",
    data() {
        return {
            items: [],
            fields: [
                "id",
                "name",
                "type",
                { key: "weapon_name", label: "Weapon" },
                { key: "skin_style_id", label: "Skin ID" },
                { key: "skin_style_name", label: "Skin" },
                { key: "asset_id", label: "Asset ID" },
                { key: "is_default", label: "Default" },
                "dev_owned",
                { key: "gr_ready", label: "Ready" },
            ],
            modal: false,
            editing: false,
            form: blankForm(),
            typeOptions: [
                { value: 3, label: "Avatar" },
                { value: 2, label: "Weapon" },
            ],
            yesNo: [
                { value: 0, label: "No" },
                { value: 1, label: "Yes" },
            ],
            skinOptions: [],
            weaponOptions: [],
            toastMsg: [],
            toastColor: [],
            fixedToasts: 0,
        };
    },
    mounted() {
        this.load();
        this.loadSkins();
        this.loadWeapons();
    },
    methods: {
        toast(msg, color) {
            this.fixedToasts++;
            this.toastMsg[this.fixedToasts] = msg;
            this.toastColor[this.fixedToasts] = color || "primary";
        },
        load() {
            this.axios.get(this.$gstate.getBaseUrl() + "/item/admin/all", this.$auth.apiOptions)
                .then(response => { this.items = response.data.data || []; })
                .catch(error => this.fail(error));
        },
        loadSkins() {
            this.axios.get(this.$gstate.getBaseUrl() + "/skin/admin/all", this.$auth.apiOptions)
                .then(response => {
                    this.skinOptions = (response.data.data || []).map(skin => ({
                        value: skin.id,
                        label: skin.name + " - " + skin.id,
                    }));
                })
                .catch(error => this.fail(error));
        },
        loadWeapons() {
            this.axios.get(this.$gstate.getBaseUrl() + "/item/admin/weapons", this.$auth.apiOptions)
                .then(response => {
                    this.weaponOptions = (response.data.data || []).map(weapon => ({
                        value: weapon.id,
                        label: weapon.name + " - " + weapon.id,
                    }));
                })
                .catch(error => this.fail(error));
        },
        openCreate() {
            this.editing = false;
            this.form = blankForm();
            this.modal = true;
        },
        openEdit(item) {
            this.editing = true;
            this.form = {
                id: item.id,
                name: item.name,
                type: item.type_id,
                weapon_type: item.weapon_type || "",
                skin_style_id: item.skin_style_id,
                asset_id: item.asset_id,
                is_default: item.is_default,
                dev_owned: item.dev_owned,
                can_own_multiple: item.can_own_multiple,
                gr_ready: item.gr_ready,
            };
            this.modal = true;
        },
        save() {
            if (!this.form.name || !this.form.asset_id || !this.form.skin_style_id) {
                this.toast("Name, asset ID, and skin style are required.", "danger");
                return;
            }
            if (this.form.type == 2 && !this.form.weapon_type) {
                this.toast("Weapon type is required for a gun skin.", "danger");
                return;
            }
            const data = new FormData();
            data.append("name", this.form.name);
            data.append("asset_id", this.form.asset_id);
            data.append("skin_style_id", this.form.skin_style_id);
            data.append("weapon_type", this.form.type == 2 ? this.form.weapon_type : "");
            data.append("is_default", this.form.is_default);
            data.append("dev_owned", this.form.dev_owned);
            data.append("can_own_multiple", this.form.can_own_multiple);
            data.append("gr_ready", this.form.gr_ready);
            let url = this.$gstate.getBaseUrl() + "/item/admin/create";
            if (this.editing) {
                url = this.$gstate.getBaseUrl() + "/item/admin/update/" + this.form.id;
            } else {
                data.append("type", this.form.type);
            }
            this.axios.post(url, data, this.$auth.apiOptions)
                .then(response => {
                    this.modal = false;
                    this.toast((response.data.messages && response.data.messages[0]) || "Saved", "success");
                    this.load();
                })
                .catch(error => this.fail(error));
        },
        fail(error) {
            console.log(error);
            const data = error && error.response && error.response.data;
            this.toast((data && (data.error || (data.messages && data.messages[0]))) || error, "danger");
        },
    },
};
</script>
