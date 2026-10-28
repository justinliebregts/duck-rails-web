import { reactive, watch } from '@vue/composition-api';
import { useListQuery } from '@/models/useListQuery.js';

export function useBanListQuery(fetchFn, defaultLimit = 100, defaultType = 'game') {
    // 1. 取得基礎列表查詢狀態與方法
    const baseList = useListQuery(fetchFn, defaultLimit);

    // 2. 定義 Ban 列表特有的額外狀態
    const banState = reactive({
        active_only: false,
        ban_type: defaultType
    });


    // 3. 監聽 Ban 特有狀態變更（如切換過濾條件） -> 自動重置到第一頁並觸發 API
    watch(
        () => [banState.active_only, banState.ban_type],
        ([newActive, newType], [oldActive, oldType]) => {
            // console.log("[banState]", banState, newActive, newType, oldActive, oldType);
            if (newActive !== oldActive || newType !== oldType) {
                baseList.submitAndFetch();
            }
        }
    );

    // 4. 擴充重置方法（同時清空基礎與 Ban 狀態）
    const reset = () => {
        baseList.reset();
        banState.active_only = false;
        banState.ban_type = defaultType;
    };

    // 5. 合併兩者的參數輸出
    const toParams = () => ({
        ...baseList.toParams(),
        ...banState
    });

    return {
        query: baseList.query, // 對應模板的 banQuery.query.q 等
        banState,              // 對應模板的 banQuery.banState.ban_type 等

        //redeclared
        reset,
        toParams,

        //redirect from baseList
        setPageSize: baseList.setPageSize,
        resetAndFetch: baseList.resetAndFetch,
        submitAndFetch: baseList.submitAndFetch,
    };
}