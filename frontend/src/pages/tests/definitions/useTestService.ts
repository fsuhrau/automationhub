import {useMemo} from 'react';
import {useApiClient} from '../../../api/ApiClientProvider';
import {testService, TestService} from '../../../services/test.service';

// useTestService is the test API on the injected ApiClient
export const useTestService = (): TestService => {
    const api = useApiClient();
    return useMemo(() => testService(api), [api]);
};
