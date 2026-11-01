<template>
    <CCard>
        <CCardHeader>
            <strong>Duck Coins</strong>
            <CButton size="sm" color="light" class="float-right" :disabled="loading" @click="load">Refresh</CButton>
        </CCardHeader>
        <CCardBody>
            <CRow>
                <CCol lg="3">
                    Balance: <b>{{ wallet ? $gstate.numberWithCommas(wallet.coins) : '…' }}</b> coins<br />
                    <template v-if="wallet">
                        Store purchases: <small><b>{{ wallet.purchases.join(', ') || 'none' }}</b></small><br />
                        IAP entitlements: <small><b>{{ wallet.entitlements.join(', ') || 'none' }}</b></small><br />
                    </template>
                </CCol>
                <CCol lg="9">
                    <CRow class="align-items-end">
                        <CCol sm="3">
                            <CInput
                                label="Amount"
                                type="number"
                                :value.sync="amount"
                                description="Negative removes coins"
                            />
                        </CCol>
                        <CCol sm="3" class="mb-3">
                            <CButton color="success" class="mr-2" :disabled="busy" @click="confirmOpen = true">
                                {{ isRemove ? 'Remove' : 'Grant' }}
                            </CButton>
                        </CCol>
                    </CRow>
                </CCol>
            </CRow>

            <CDataTable
                v-if="wallet && wallet.adjustments.length"
                :items="wallet.adjustments"
                :fields="adjustmentFields"
                small
                striped
                class="mt-3"
            >
                <template #amount="{ item }">
                    <td :class="item.amount < 0 ? 'text-danger' : 'text-success'">
                        {{ item.amount > 0 ? '+' : '' }}{{ $gstate.numberWithCommas(item.amount) }}
                    </td>
                </template>
            </CDataTable>
        </CCardBody>

        <CModal title="Confirm duck coin adjustment" :show.sync="confirmOpen" color="warning">
            {{ isRemove ? 'Remove' : 'Grant' }}
            <b>{{ $gstate.numberWithCommas(Math.abs(parsedAmount || 0)) }}</b> duck coins
            {{ isRemove ? 'from' : 'to' }} <b>{{ playerId }}</b>?
            <template #footer>
                <CButton color="secondary" @click="confirmOpen = false">Cancel</CButton>
                <CButton color="warning" :disabled="busy" @click="submit">Confirm</CButton>
            </template>
        </CModal>
    </CCard>
</template>

<script>
export default {
    name: 'DuckCoinsCard',
    props: {
        playerId: { type: String, required: true },
    },
    data() {
        return {
            wallet: null,
            loading: false,
            busy: false,
            amount: '',
            confirmOpen: false,
            adjustmentFields: [
                { key: 'created_at', label: 'When' },
                { key: 'amount', label: 'Amount' },
                { key: 'admin_id', label: 'Admin' },
            ],
        };
    },
    computed: {
        parsedAmount() {
            return parseInt(this.amount, 10);
        },
        isRemove() {
            return this.parsedAmount < 0;
        },
    },
    watch: {
        playerId() {
            this.load();
        },
    },
    mounted() {
        this.load();
    },
    methods: {
        errorMessage(error, fallback) {
            const data = error && error.response && error.response.data;
            return (data && (data.error || (data.messages && data.messages[0]))) || fallback;
        },
        load() {
            if (!this.playerId) return;
            this.loading = true;
            const url = this.$gstate.getBaseUrl() + '/admin/wallet/' + this.playerId;
            this.axios.get(url, this.$auth.apiOptions)
                .then((response) => {
                    this.wallet = response.data.data;
                })
                .catch((error) => {
                    console.log(error);
                    this.$emit('toast', this.errorMessage(error, 'Failed to load duck coins'), 'danger');
                })
                .finally(() => {
                    this.loading = false;
                });
        },
        submit() {
            const amount = this.parsedAmount;
            if (!Number.isFinite(amount) || amount === 0 || Math.abs(amount) > 1000000) {
                this.$emit('toast', 'Enter a non-zero amount between -1,000,000 and 1,000,000', 'warning');
                this.confirmOpen = false;
                return;
            }
            this.busy = true;
            const url = this.$gstate.getBaseUrl() + '/admin/coins/adjust';
            const body = {
                player_id: this.playerId,
                amount: amount,
            };
            this.axios.post(url, body, this.$auth.apiJsonOptions || this.$auth.apiOptions)
                .then((response) => {
                    const data = response.data.data;
                    const verb = amount < 0 ? 'Removed' : 'Granted';
                    this.$emit('toast', `${verb} ${Math.abs(amount)} coins. New balance: ${data.coins}`, 'success');
                    this.amount = '';
                    this.confirmOpen = false;
                    this.load();
                })
                .catch((error) => {
                    console.log(error);
                    this.$emit('toast', this.errorMessage(error, 'Duck coin adjustment failed'), 'danger');
                })
                .finally(() => {
                    this.busy = false;
                });
        },
    },
};
</script>
