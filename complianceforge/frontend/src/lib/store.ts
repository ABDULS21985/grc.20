import { create } from 'zustand';
import type { User, DashboardSummary, ComplianceFramework, ComplianceScore } from '@/types';

// ── Auth Store ─────────────────────────────────────────
interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  setUser: (user: User, token: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== 'undefined' ? localStorage.getItem('cf_token') : null,
  isAuthenticated: typeof window !== 'undefined' ? !!localStorage.getItem('cf_token') : false,

  setUser: (user, token) => {
    localStorage.setItem('cf_token', token);
    set({ user, token, isAuthenticated: true });
  },

  logout: () => {
    localStorage.removeItem('cf_token');
    set({ user: null, token: null, isAuthenticated: false });
    window.location.href = '/auth/login';
  },
}));

// ── Dashboard Store ────────────────────────────────────
interface DashboardState {
  summary: DashboardSummary | null;
  loading: boolean;
  error: string | null;
  lastFetched: number | null;
  setSummary: (summary: DashboardSummary) => void;
  setLoading: (loading: boolean) => void;
  setError: (error: string | null) => void;
  shouldRefresh: () => boolean;
}

export const useDashboardStore = create<DashboardState>((set, get) => ({
  summary: null,
  loading: false,
  error: null,
  lastFetched: null,

  setSummary: (summary) => set({ summary, lastFetched: Date.now(), error: null }),
  setLoading: (loading) => set({ loading }),
  setError: (error) => set({ error, loading: false }),

  shouldRefresh: () => {
    const { lastFetched } = get();
    if (!lastFetched) return true;
    return Date.now() - lastFetched > 2 * 60 * 1000; // 2 minutes
  },
}));

// ── Frameworks Store ───────────────────────────────────
interface FrameworksState {
  frameworks: ComplianceFramework[];
  scores: ComplianceScore[];
  selectedFrameworkId: string | null;
  loading: boolean;
  setFrameworks: (frameworks: ComplianceFramework[]) => void;
  setScores: (scores: ComplianceScore[]) => void;
  selectFramework: (id: string | null) => void;
  setLoading: (loading: boolean) => void;
}

export const useFrameworksStore = create<FrameworksState>((set) => ({
  frameworks: [],
  scores: [],
  selectedFrameworkId: null,
  loading: false,

  setFrameworks: (frameworks) => set({ frameworks }),
  setScores: (scores) => set({ scores }),
  selectFramework: (id) => set({ selectedFrameworkId: id }),
  setLoading: (loading) => set({ loading }),
}));

// ── Notification Store ─────────────────────────────────
interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  timestamp: number;
}

interface NotificationState {
  notifications: Notification[];
  add: (type: Notification['type'], title: string, message: string) => void;
  dismiss: (id: string) => void;
  clear: () => void;
}

export const useNotificationStore = create<NotificationState>((set, get) => ({
  notifications: [],

  add: (type, title, message) => {
    const id = Math.random().toString(36).substring(7);
    const notification: Notification = { id, type, title, message, timestamp: Date.now() };
    set({ notifications: [...get().notifications, notification] });
    // Auto-dismiss after 5 seconds
    setTimeout(() => {
      set({ notifications: get().notifications.filter(n => n.id !== id) });
    }, 5000);
  },

  dismiss: (id) => set({ notifications: get().notifications.filter(n => n.id !== id) }),
  clear: () => set({ notifications: [] }),
}));
