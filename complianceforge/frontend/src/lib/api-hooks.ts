// ComplianceForge — Custom React Hooks for API
// Lightweight data-fetching hooks built on top of the api client.
// No external query library needed — uses useState/useEffect/useCallback.

import { useState, useEffect, useCallback, useRef } from 'react';
import api from './api';

// ── Generic Query Hook ──────────────────────────────────

export interface UseQueryResult<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

function useApiQuery<T>(
  fetcher: () => Promise<any>,
  deps: any[] = []
): UseQueryResult<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const mountedRef = useRef(true);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await fetcher();
      if (mountedRef.current) {
        // The API returns { success, data, ... } — unwrap if present
        setData(response?.data !== undefined ? response.data : response);
      }
    } catch (err: any) {
      if (mountedRef.current) {
        setError(err?.message || 'An unexpected error occurred');
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  useEffect(() => {
    mountedRef.current = true;
    fetchData();
    return () => {
      mountedRef.current = false;
    };
  }, [fetchData]);

  return { data, loading, error, refetch: fetchData };
}

// ── Generic Mutation Hook ───────────────────────────────

export interface UseMutationResult<TInput, TOutput = any> {
  mutate: (data: TInput) => Promise<TOutput>;
  loading: boolean;
  error: string | null;
}

function useApiMutation<TInput, TOutput = any>(
  mutator: (data: TInput) => Promise<any>
): UseMutationResult<TInput, TOutput> {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(
    async (data: TInput): Promise<TOutput> => {
      setLoading(true);
      setError(null);
      try {
        const response = await mutator(data);
        return response?.data !== undefined ? response.data : response;
      } catch (err: any) {
        const message = err?.message || 'An unexpected error occurred';
        setError(message);
        throw err;
      } finally {
        setLoading(false);
      }
    },
    [mutator]
  );

  return { mutate, loading, error };
}

// A variant for mutations that take no input data (just an ID already bound in the closure).
function useApiAction<TOutput = any>(
  action: () => Promise<any>
): { execute: () => Promise<TOutput>; loading: boolean; error: string | null } {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const execute = useCallback(async (): Promise<TOutput> => {
    setLoading(true);
    setError(null);
    try {
      const response = await action();
      return response?.data !== undefined ? response.data : response;
    } catch (err: any) {
      const message = err?.message || 'An unexpected error occurred';
      setError(message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [action]);

  return { execute, loading, error };
}

// ── Dashboard ───────────────────────────────────────────

export function useDashboard() {
  return useApiQuery<any>(() => api.getDashboard());
}

// ── Frameworks ──────────────────────────────────────────

export function useFrameworks() {
  return useApiQuery<any>(() => api.getFrameworks());
}

export function useFramework(id: string) {
  return useApiQuery<any>(() => api.getFramework(id), [id]);
}

export function useFrameworkControls(id: string, page = 1) {
  return useApiQuery<any>(() => api.getFrameworkControls(id, page), [id, page]);
}

// ── Compliance ──────────────────────────────────────────

export function useComplianceScores() {
  return useApiQuery<any>(() => api.getComplianceScores());
}

export function useGapAnalysis(frameworkId?: string) {
  return useApiQuery<any>(
    () => api.getGapAnalysis(frameworkId),
    [frameworkId]
  );
}

export function useCrossMapping() {
  return useApiQuery<any>(() => api.getCrossMapping());
}

// ── Risks ───────────────────────────────────────────────

export function useRisks(page = 1, sortBy = 'residual_risk_score', sortDir = 'desc') {
  return useApiQuery<any>(
    () => api.getRisks(page, 20, sortBy, sortDir),
    [page, sortBy, sortDir]
  );
}

export function useRisk(id: string) {
  return useApiQuery<any>(() => api.getRisk(id), [id]);
}

export function useRiskHeatmap() {
  return useApiQuery<any>(() => api.getRiskHeatmap());
}

export function useCreateRisk() {
  return useApiMutation<any>((data) => api.createRisk(data));
}

// ── Policies ────────────────────────────────────────────

export function usePolicies(page = 1) {
  return useApiQuery<any>(() => api.getPolicies(page), [page]);
}

export function usePolicy(id: string) {
  return useApiQuery<any>(() => api.getPolicy(id), [id]);
}

export function useAttestationStats() {
  return useApiQuery<any>(() => api.getAttestationStats());
}

export function useCreatePolicy() {
  return useApiMutation<any>((data) => api.createPolicy(data));
}

export function usePublishPolicy() {
  return useApiMutation<string>((id) => api.publishPolicy(id));
}

// ── Audits ──────────────────────────────────────────────

export function useAudits(page = 1) {
  return useApiQuery<any>(() => api.getAudits(page), [page]);
}

export function useAudit(id: string) {
  return useApiQuery<any>(() => api.getAudit(id), [id]);
}

export function useAuditFindings(auditId: string) {
  return useApiQuery<any>(() => api.getAuditFindings(auditId), [auditId]);
}

export function useFindingsStats() {
  return useApiQuery<any>(() => api.getFindingsStats());
}

export function useCreateAudit() {
  return useApiMutation<any>((data) => api.createAudit(data));
}

export function useCreateFinding(auditId: string) {
  return useApiMutation<any>((data) => api.createFinding(auditId, data));
}

// ── Incidents ───────────────────────────────────────────

export function useIncidents(page = 1) {
  return useApiQuery<any>(() => api.getIncidents(page), [page]);
}

export function useIncident(id: string) {
  return useApiQuery<any>(() => api.getIncident(id), [id]);
}

export function useIncidentStats() {
  return useApiQuery<any>(() => api.getIncidentStats());
}

export function useUrgentBreaches() {
  return useApiQuery<any>(() => api.getUrgentBreaches());
}

export function useReportIncident() {
  return useApiMutation<any>((data) => api.reportIncident(data));
}

export function useNotifyDPA() {
  return useApiMutation<string>((id) => api.notifyDPA(id));
}

// ── Vendors ─────────────────────────────────────────────

export function useVendors(page = 1) {
  return useApiQuery<any>(() => api.getVendors(page), [page]);
}

export function useVendor(id: string) {
  return useApiQuery<any>(() => api.getVendor(id), [id]);
}

export function useVendorStats() {
  return useApiQuery<any>(() => api.getVendorStats());
}

export function useOnboardVendor() {
  return useApiMutation<any>((data) => api.onboardVendor(data));
}

// ── Assets ──────────────────────────────────────────────

export function useAssets(page = 1) {
  return useApiQuery<any>(() => api.getAssets(page), [page]);
}

export function useAssetStats() {
  return useApiQuery<any>(() => api.getAssetStats());
}

export function useRegisterAsset() {
  return useApiMutation<any>((data) => api.registerAsset(data));
}

// ── Settings ────────────────────────────────────────────

export function useOrganization() {
  return useApiQuery<any>(() => api.getOrganization());
}

export function useUsers(page = 1, search = '') {
  return useApiQuery<any>(
    () => api.getUsers(page, search),
    [page, search]
  );
}

export function useRoles() {
  return useApiQuery<any>(() => api.getRoles());
}

export function useAuditLog(page = 1) {
  return useApiQuery<any>(() => api.getAuditLog(page), [page]);
}

export function useCreateUser() {
  return useApiMutation<any>((data) => api.createUser(data));
}
