'use client';

import { useEffect, useState, useCallback } from 'react';
import api from '@/lib/api';

// ============================================================
// TYPES
// ============================================================

interface CalendarEvent {
  id: string;
  event_ref: string;
  event_type: string;
  category: string;
  priority: string;
  status: string;
  title: string;
  description: string;
  source_entity_type: string;
  source_entity_id: string;
  source_entity_ref: string;
  due_date: string;
  all_day: boolean;
  recurrence_type: string;
  assigned_to_user_id: string | null;
  assigned_to_name: string;
  owner_name: string;
  escalated: boolean;
  rescheduled_count: number;
  tags: string[];
  created_at: string;
}

interface UpcomingDeadline {
  id: string;
  event_ref: string;
  event_type: string;
  category: string;
  priority: string;
  status: string;
  title: string;
  due_date: string;
  days_left: number;
  source_entity_type: string;
  source_entity_ref: string;
  assigned_to_name: string;
}

interface OverdueItem {
  id: string;
  event_ref: string;
  event_type: string;
  category: string;
  priority: string;
  title: string;
  due_date: string;
  days_overdue: number;
  source_entity_ref: string;
  assigned_to_name: string;
  escalated: boolean;
}

interface CalendarSummary {
  month: string;
  total_events: number;
  upcoming_count: number;
  due_soon_count: number;
  overdue_count: number;
  completed_count: number;
  cancelled_count: number;
  by_category: Record<string, number>;
  by_priority: Record<string, number>;
  completion_rate: number;
  critical_deadlines: UpcomingDeadline[];
  weekly_summary: WeekSummary[];
}

interface WeekSummary {
  week_start: string;
  week_end: string;
  total: number;
  completed: number;
  overdue: number;
}

interface SyncStatus {
  modules: SyncModule[];
  last_full_sync: string | null;
  total_events: number;
  overdue_events: number;
}

interface SyncModule {
  module_name: string;
  is_enabled: boolean;
  last_sync_at: string | null;
  last_sync_status: string;
  last_sync_events_created: number;
  last_sync_events_updated: number;
}

// ============================================================
// CONSTANTS
// ============================================================

const PRIORITY_COLORS: Record<string, string> = {
  critical: 'bg-red-100 text-red-700 border-red-300',
  high: 'bg-orange-100 text-orange-700 border-orange-300',
  medium: 'bg-amber-100 text-amber-700 border-amber-300',
  low: 'bg-blue-100 text-blue-700 border-blue-300',
};

const STATUS_COLORS: Record<string, string> = {
  upcoming: 'bg-gray-100 text-gray-700',
  due_soon: 'bg-yellow-100 text-yellow-700',
  overdue: 'bg-red-100 text-red-700',
  completed: 'bg-green-100 text-green-700',
  cancelled: 'bg-gray-100 text-gray-400',
  snoozed: 'bg-purple-100 text-purple-700',
};

const CATEGORY_COLORS: Record<string, string> = {
  policy: 'bg-indigo-500',
  risk: 'bg-red-500',
  vendor: 'bg-emerald-500',
  audit: 'bg-purple-500',
  evidence: 'bg-sky-500',
  exception: 'bg-orange-500',
  dsr: 'bg-pink-500',
  incident: 'bg-rose-500',
  regulatory: 'bg-teal-500',
  business_continuity: 'bg-cyan-500',
  board: 'bg-violet-500',
  custom: 'bg-gray-500',
};

const CATEGORY_LABELS: Record<string, string> = {
  policy: 'Policy',
  risk: 'Risk',
  vendor: 'Vendor',
  audit: 'Audit',
  evidence: 'Evidence',
  exception: 'Exception',
  dsr: 'DSR',
  incident: 'Incident',
  regulatory: 'Regulatory',
  business_continuity: 'Business Continuity',
  board: 'Board',
  custom: 'Custom',
};

const VIEW_MODES = ['month', 'week', 'agenda'] as const;
type ViewMode = (typeof VIEW_MODES)[number];

// ============================================================
// HELPERS
// ============================================================

function formatDate(dateStr: string): string {
  const date = new Date(dateStr + 'T00:00:00');
  return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
}

function formatEventType(type: string): string {
  return type.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
}

function getDaysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

function getFirstDayOfMonth(year: number, month: number): number {
  return new Date(year, month, 1).getDay();
}

// ============================================================
// COMPONENT
// ============================================================

export default function CalendarPage() {
  const [viewMode, setViewMode] = useState<ViewMode>('month');
  const [currentDate, setCurrentDate] = useState(new Date());
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [deadlines, setDeadlines] = useState<UpcomingDeadline[]>([]);
  const [overdueItems, setOverdueItems] = useState<OverdueItem[]>([]);
  const [summary, setSummary] = useState<CalendarSummary | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState(false);
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [selectedPriority, setSelectedPriority] = useState<string>('');
  const [showSidebar, setShowSidebar] = useState(true);

  const currentYear = currentDate.getFullYear();
  const currentMonth = currentDate.getMonth();
  const monthStr = `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}`;

  // --------------------------------------------------------
  // DATA LOADING
  // --------------------------------------------------------

  const loadData = useCallback(async () => {
    setLoading(true);
    try {
      const startDate = new Date(currentYear, currentMonth, 1).toISOString().split('T')[0];
      const endDate = new Date(currentYear, currentMonth + 1, 0).toISOString().split('T')[0];

      let eventsUrl = `/calendar/events?start_date=${startDate}&end_date=${endDate}&page_size=100`;
      if (selectedCategory) eventsUrl += `&category=${selectedCategory}`;
      if (selectedPriority) eventsUrl += `&priority=${selectedPriority}`;

      const [eventsRes, deadlinesRes, overdueRes, summaryRes, syncRes] = await Promise.allSettled([
        api.get<{ data: CalendarEvent[] }>(eventsUrl),
        api.get<{ data: UpcomingDeadline[] }>('/calendar/deadlines?days=30&limit=20'),
        api.get<{ data: OverdueItem[] }>('/calendar/overdue'),
        api.get<{ data: CalendarSummary }>(`/calendar/summary?month=${monthStr}`),
        api.get<{ data: SyncStatus }>('/calendar/sync/status'),
      ]);

      if (eventsRes.status === 'fulfilled') setEvents(eventsRes.value.data || []);
      if (deadlinesRes.status === 'fulfilled') setDeadlines(deadlinesRes.value.data || []);
      if (overdueRes.status === 'fulfilled') setOverdueItems(overdueRes.value.data || []);
      if (summaryRes.status === 'fulfilled') setSummary(summaryRes.value.data || null);
      if (syncRes.status === 'fulfilled') setSyncStatus(syncRes.value.data || null);
    } catch (err) {
      console.error('Failed to load calendar data:', err);
    } finally {
      setLoading(false);
    }
  }, [currentYear, currentMonth, monthStr, selectedCategory, selectedPriority]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // --------------------------------------------------------
  // ACTIONS
  // --------------------------------------------------------

  const handleSync = async () => {
    setSyncing(true);
    try {
      await api.post('/calendar/sync/trigger', {});
      await loadData();
    } catch (err) {
      console.error('Sync failed:', err);
    } finally {
      setSyncing(false);
    }
  };

  const handleComplete = async (eventId: string) => {
    try {
      await api.put(`/calendar/events/${eventId}/complete`, {});
      await loadData();
    } catch (err) {
      console.error('Failed to complete event:', err);
    }
  };

  const navigateMonth = (direction: number) => {
    setCurrentDate(new Date(currentYear, currentMonth + direction, 1));
  };

  // --------------------------------------------------------
  // MONTH VIEW
  // --------------------------------------------------------

  const getEventsForDate = (dateStr: string) => {
    return events.filter(e => e.due_date === dateStr);
  };

  const renderMonthView = () => {
    const daysInMonth = getDaysInMonth(currentYear, currentMonth);
    const firstDay = getFirstDayOfMonth(currentYear, currentMonth);
    const weeks: (number | null)[][] = [];
    let currentWeek: (number | null)[] = [];

    // Fill leading empty days
    for (let i = 0; i < firstDay; i++) {
      currentWeek.push(null);
    }

    for (let day = 1; day <= daysInMonth; day++) {
      currentWeek.push(day);
      if (currentWeek.length === 7) {
        weeks.push(currentWeek);
        currentWeek = [];
      }
    }

    // Fill trailing empty days
    if (currentWeek.length > 0) {
      while (currentWeek.length < 7) {
        currentWeek.push(null);
      }
      weeks.push(currentWeek);
    }

    const today = new Date();
    const todayStr = today.toISOString().split('T')[0];

    return (
      <div className="border rounded-lg overflow-hidden">
        {/* Day headers */}
        <div className="grid grid-cols-7 bg-gray-50 border-b">
          {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map(day => (
            <div key={day} className="px-2 py-2 text-center text-xs font-medium text-gray-500 uppercase">
              {day}
            </div>
          ))}
        </div>

        {/* Calendar grid */}
        {weeks.map((week, wi) => (
          <div key={wi} className="grid grid-cols-7 border-b last:border-b-0">
            {week.map((day, di) => {
              if (day === null) {
                return <div key={di} className="min-h-[100px] bg-gray-50 border-r last:border-r-0" />;
              }

              const dateStr = `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
              const dayEvents = getEventsForDate(dateStr);
              const isToday = dateStr === todayStr;

              return (
                <div
                  key={di}
                  className={`min-h-[100px] border-r last:border-r-0 p-1 ${isToday ? 'bg-blue-50' : 'bg-white'}`}
                >
                  <div className={`text-sm font-medium mb-1 ${isToday ? 'text-blue-600' : 'text-gray-700'}`}>
                    {day}
                  </div>
                  <div className="space-y-0.5">
                    {dayEvents.slice(0, 3).map(event => (
                      <div
                        key={event.id}
                        className={`text-xs px-1 py-0.5 rounded truncate cursor-pointer ${CATEGORY_COLORS[event.category] || 'bg-gray-500'} text-white`}
                        title={`${event.title} [${event.priority.toUpperCase()}]`}
                      >
                        {event.title}
                      </div>
                    ))}
                    {dayEvents.length > 3 && (
                      <div className="text-xs text-gray-500 px-1">+{dayEvents.length - 3} more</div>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        ))}
      </div>
    );
  };

  // --------------------------------------------------------
  // WEEK VIEW
  // --------------------------------------------------------

  const renderWeekView = () => {
    const startOfWeek = new Date(currentDate);
    startOfWeek.setDate(currentDate.getDate() - currentDate.getDay());

    const days: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const d = new Date(startOfWeek);
      d.setDate(startOfWeek.getDate() + i);
      days.push(d);
    }

    const todayStr = new Date().toISOString().split('T')[0];

    return (
      <div className="border rounded-lg overflow-hidden">
        <div className="grid grid-cols-7">
          {days.map(day => {
            const dateStr = day.toISOString().split('T')[0];
            const dayEvents = getEventsForDate(dateStr);
            const isToday = dateStr === todayStr;

            return (
              <div
                key={dateStr}
                className={`min-h-[400px] border-r last:border-r-0 ${isToday ? 'bg-blue-50' : 'bg-white'}`}
              >
                <div className={`px-2 py-2 border-b text-center ${isToday ? 'bg-blue-100' : 'bg-gray-50'}`}>
                  <div className="text-xs text-gray-500 uppercase">
                    {day.toLocaleDateString('en-GB', { weekday: 'short' })}
                  </div>
                  <div className={`text-lg font-semibold ${isToday ? 'text-blue-600' : 'text-gray-900'}`}>
                    {day.getDate()}
                  </div>
                </div>
                <div className="p-1 space-y-1">
                  {dayEvents.map(event => (
                    <div
                      key={event.id}
                      className={`text-xs p-1.5 rounded border ${PRIORITY_COLORS[event.priority] || 'bg-gray-100'}`}
                    >
                      <div className="font-medium truncate">{event.title}</div>
                      <div className="text-[10px] mt-0.5 opacity-75">
                        {CATEGORY_LABELS[event.category] || event.category}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  // --------------------------------------------------------
  // AGENDA VIEW
  // --------------------------------------------------------

  const renderAgendaView = () => {
    const sortedEvents = [...events].sort((a, b) => a.due_date.localeCompare(b.due_date));

    // Group by date
    const grouped: Record<string, CalendarEvent[]> = {};
    for (const event of sortedEvents) {
      if (!grouped[event.due_date]) grouped[event.due_date] = [];
      grouped[event.due_date].push(event);
    }

    const dates = Object.keys(grouped).sort();

    if (dates.length === 0) {
      return (
        <div className="text-center py-12 text-gray-500">
          No events found for this period.
        </div>
      );
    }

    return (
      <div className="space-y-4">
        {dates.map(date => (
          <div key={date} className="border rounded-lg overflow-hidden">
            <div className="bg-gray-50 px-4 py-2 border-b">
              <span className="font-medium text-gray-900">{formatDate(date)}</span>
              <span className="ml-2 text-sm text-gray-500">
                ({grouped[date].length} event{grouped[date].length !== 1 ? 's' : ''})
              </span>
            </div>
            <div className="divide-y">
              {grouped[date].map(event => (
                <div key={event.id} className="px-4 py-3 flex items-center justify-between hover:bg-gray-50">
                  <div className="flex items-center gap-3 min-w-0 flex-1">
                    <div className={`w-3 h-3 rounded-full flex-shrink-0 ${CATEGORY_COLORS[event.category] || 'bg-gray-500'}`} />
                    <div className="min-w-0">
                      <div className="font-medium text-gray-900 truncate">{event.title}</div>
                      <div className="text-sm text-gray-500">
                        {CATEGORY_LABELS[event.category]} &middot; {formatEventType(event.event_type)}
                        {event.assigned_to_name && <span> &middot; {event.assigned_to_name}</span>}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 flex-shrink-0">
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${PRIORITY_COLORS[event.priority]}`}>
                      {event.priority}
                    </span>
                    <span className={`px-2 py-0.5 text-xs font-medium rounded ${STATUS_COLORS[event.status]}`}>
                      {event.status.replace(/_/g, ' ')}
                    </span>
                    {event.status !== 'completed' && event.status !== 'cancelled' && (
                      <button
                        onClick={() => handleComplete(event.id)}
                        className="px-2 py-1 text-xs bg-green-500 text-white rounded hover:bg-green-600"
                      >
                        Complete
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    );
  };

  // --------------------------------------------------------
  // SIDEBAR: Deadline Dashboard
  // --------------------------------------------------------

  const renderSidebar = () => (
    <div className="w-80 space-y-4 flex-shrink-0">
      {/* Summary Card */}
      {summary && (
        <div className="border rounded-lg p-4 bg-white">
          <h3 className="text-sm font-semibold text-gray-900 mb-3">Monthly Summary</h3>
          <div className="grid grid-cols-2 gap-2">
            <div className="text-center p-2 bg-gray-50 rounded">
              <div className="text-lg font-bold text-gray-900">{summary.total_events}</div>
              <div className="text-xs text-gray-500">Total</div>
            </div>
            <div className="text-center p-2 bg-green-50 rounded">
              <div className="text-lg font-bold text-green-700">{summary.completed_count}</div>
              <div className="text-xs text-gray-500">Completed</div>
            </div>
            <div className="text-center p-2 bg-yellow-50 rounded">
              <div className="text-lg font-bold text-yellow-700">{summary.due_soon_count}</div>
              <div className="text-xs text-gray-500">Due Soon</div>
            </div>
            <div className="text-center p-2 bg-red-50 rounded">
              <div className="text-lg font-bold text-red-700">{summary.overdue_count}</div>
              <div className="text-xs text-gray-500">Overdue</div>
            </div>
          </div>
          <div className="mt-3">
            <div className="flex justify-between text-xs text-gray-500 mb-1">
              <span>Completion Rate</span>
              <span>{summary.completion_rate.toFixed(1)}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div
                className="bg-green-500 h-2 rounded-full"
                style={{ width: `${Math.min(summary.completion_rate, 100)}%` }}
              />
            </div>
          </div>
        </div>
      )}

      {/* Overdue Items */}
      {overdueItems.length > 0 && (
        <div className="border border-red-200 rounded-lg p-4 bg-red-50">
          <h3 className="text-sm font-semibold text-red-800 mb-2">
            Overdue ({overdueItems.length})
          </h3>
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {overdueItems.slice(0, 8).map(item => (
              <div key={item.id} className="bg-white rounded p-2 border border-red-200">
                <div className="text-xs font-medium text-gray-900 truncate">{item.title}</div>
                <div className="flex items-center gap-1 mt-1">
                  <span className={`px-1.5 py-0.5 text-[10px] rounded ${PRIORITY_COLORS[item.priority]}`}>
                    {item.priority}
                  </span>
                  <span className="text-[10px] text-red-600 font-medium">
                    {item.days_overdue}d overdue
                  </span>
                  {item.escalated && (
                    <span className="text-[10px] text-orange-600 font-medium">ESCALATED</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Upcoming Deadlines */}
      <div className="border rounded-lg p-4 bg-white">
        <h3 className="text-sm font-semibold text-gray-900 mb-2">
          Upcoming Deadlines (30d)
        </h3>
        {deadlines.length === 0 ? (
          <div className="text-sm text-gray-400 py-2">No upcoming deadlines</div>
        ) : (
          <div className="space-y-2 max-h-72 overflow-y-auto">
            {deadlines.map(d => (
              <div key={d.id} className="flex items-start gap-2 py-1.5 border-b last:border-b-0">
                <div className={`w-2 h-2 rounded-full mt-1.5 flex-shrink-0 ${CATEGORY_COLORS[d.category]}`} />
                <div className="min-w-0 flex-1">
                  <div className="text-xs font-medium text-gray-900 truncate">{d.title}</div>
                  <div className="text-[10px] text-gray-500">
                    {formatDate(d.due_date)} &middot;{' '}
                    <span className={d.days_left <= 3 ? 'text-red-600 font-medium' : ''}>
                      {d.days_left === 0 ? 'Today' : `${d.days_left}d left`}
                    </span>
                  </div>
                </div>
                <span className={`px-1.5 py-0.5 text-[10px] rounded flex-shrink-0 ${PRIORITY_COLORS[d.priority]}`}>
                  {d.priority}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Sync Status */}
      {syncStatus && (
        <div className="border rounded-lg p-4 bg-white">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-semibold text-gray-900">Sync Status</h3>
            <button
              onClick={handleSync}
              disabled={syncing}
              className="px-2 py-1 text-xs bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
            >
              {syncing ? 'Syncing...' : 'Sync Now'}
            </button>
          </div>
          <div className="text-xs text-gray-500 mb-2">
            {syncStatus.total_events} events total &middot; {syncStatus.overdue_events} overdue
          </div>
          <div className="space-y-1">
            {syncStatus.modules?.slice(0, 6).map(m => (
              <div key={m.module_name} className="flex items-center justify-between text-xs">
                <span className="text-gray-600 capitalize">{m.module_name.replace(/_/g, ' ')}</span>
                <span className={m.is_enabled ? 'text-green-600' : 'text-gray-400'}>
                  {m.is_enabled ? m.last_sync_status : 'disabled'}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );

  // --------------------------------------------------------
  // MAIN RENDER
  // --------------------------------------------------------

  if (loading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-64" />
          <div className="h-[500px] bg-gray-200 rounded" />
        </div>
      </div>
    );
  }

  const monthName = currentDate.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' });

  return (
    <div className="p-6 max-w-[1600px] mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Compliance Calendar</h1>
          <p className="text-sm text-gray-500 mt-1">
            Track deadlines, reviews, and compliance milestones across all modules
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowSidebar(!showSidebar)}
            className="px-3 py-2 text-sm border rounded-lg hover:bg-gray-50"
          >
            {showSidebar ? 'Hide Sidebar' : 'Show Sidebar'}
          </button>
          <button
            onClick={handleSync}
            disabled={syncing}
            className="px-4 py-2 text-sm bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            {syncing ? 'Syncing...' : 'Sync All Modules'}
          </button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-2">
          {/* Navigation */}
          <button onClick={() => navigateMonth(-1)} className="px-3 py-2 border rounded-lg hover:bg-gray-50 text-sm">
            Previous
          </button>
          <button
            onClick={() => setCurrentDate(new Date())}
            className="px-3 py-2 border rounded-lg hover:bg-gray-50 text-sm font-medium"
          >
            Today
          </button>
          <button onClick={() => navigateMonth(1)} className="px-3 py-2 border rounded-lg hover:bg-gray-50 text-sm">
            Next
          </button>
          <h2 className="text-lg font-semibold text-gray-900 ml-2">{monthName}</h2>
        </div>

        <div className="flex items-center gap-2">
          {/* Filters */}
          <select
            value={selectedCategory}
            onChange={e => setSelectedCategory(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm bg-white"
          >
            <option value="">All Categories</option>
            {Object.entries(CATEGORY_LABELS).map(([key, label]) => (
              <option key={key} value={key}>{label}</option>
            ))}
          </select>

          <select
            value={selectedPriority}
            onChange={e => setSelectedPriority(e.target.value)}
            className="px-3 py-2 border rounded-lg text-sm bg-white"
          >
            <option value="">All Priorities</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>

          {/* View Mode Toggle */}
          <div className="flex border rounded-lg overflow-hidden">
            {VIEW_MODES.map(mode => (
              <button
                key={mode}
                onClick={() => setViewMode(mode)}
                className={`px-3 py-2 text-sm capitalize ${
                  viewMode === mode
                    ? 'bg-blue-600 text-white'
                    : 'bg-white text-gray-700 hover:bg-gray-50'
                }`}
              >
                {mode}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex gap-4">
        <div className="flex-1 min-w-0">
          {viewMode === 'month' && renderMonthView()}
          {viewMode === 'week' && renderWeekView()}
          {viewMode === 'agenda' && renderAgendaView()}
        </div>
        {showSidebar && renderSidebar()}
      </div>

      {/* Category Legend */}
      <div className="mt-4 flex items-center gap-3 flex-wrap">
        {Object.entries(CATEGORY_LABELS).map(([key, label]) => (
          <div key={key} className="flex items-center gap-1.5">
            <div className={`w-2.5 h-2.5 rounded-full ${CATEGORY_COLORS[key]}`} />
            <span className="text-xs text-gray-600">{label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
