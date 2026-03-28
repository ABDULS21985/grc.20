'use client';

import { useEffect } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  LayoutDashboard,
  Shield,
  AlertTriangle,
  FileText,
  ClipboardCheck,
  AlertCircle,
  Building2,
  Package,
  Settings,
  ChevronLeft,
  ChevronRight,
  LogOut,
} from 'lucide-react';
import { cn } from '@/lib/utils';

// ── Navigation Items ─────────────────────────────────────
const NAV_ITEMS = [
  { label: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { label: 'Frameworks', href: '/frameworks', icon: Shield },
  { label: 'Risk Register', href: '/risks', icon: AlertTriangle },
  { label: 'Policies', href: '/policies', icon: FileText },
  { label: 'Audits', href: '/audits', icon: ClipboardCheck },
  { label: 'Incidents', href: '/incidents', icon: AlertCircle },
  { label: 'Vendors', href: '/vendors', icon: Building2 },
  { label: 'Assets', href: '/assets', icon: Package },
  { label: 'Settings', href: '/settings', icon: Settings },
] as const;

// ── Types ────────────────────────────────────────────────
interface SidebarUser {
  first_name: string;
  last_name: string;
  email: string;
  roles?: { name: string }[];
}

interface SidebarProps {
  collapsed: boolean;
  onToggle: () => void;
  user?: SidebarUser;
}

// ── Component ────────────────────────────────────────────
export function Sidebar({ collapsed, onToggle, user }: SidebarProps) {
  const pathname = usePathname();

  // Persist collapse state to localStorage
  useEffect(() => {
    localStorage.setItem('sidebar-collapsed', JSON.stringify(collapsed));
  }, [collapsed]);

  const initials = user
    ? `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.toUpperCase()
    : 'CF';

  const roleName = user?.roles?.[0]?.name ?? 'User';

  return (
    <aside
      className={cn(
        'flex h-screen flex-col border-r border-gray-200 bg-white transition-all duration-300',
        collapsed ? 'w-[68px]' : 'w-64',
      )}
    >
      {/* Logo */}
      <div className="flex h-16 items-center gap-3 border-b border-gray-200 px-4">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-indigo-600 text-sm font-bold text-white">
          CF
        </div>
        {!collapsed && (
          <span className="text-lg font-semibold text-gray-900 whitespace-nowrap">
            ComplianceForge
          </span>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto px-3 py-4">
        {NAV_ITEMS.map((item) => {
          const isActive =
            pathname === item.href || pathname.startsWith(`${item.href}/`);
          const Icon = item.icon;

          return (
            <Link
              key={item.href}
              href={item.href}
              title={collapsed ? item.label : undefined}
              className={cn(
                'group flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-indigo-50 text-indigo-700'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900',
              )}
            >
              <Icon
                className={cn(
                  'h-5 w-5 shrink-0',
                  isActive
                    ? 'text-indigo-600'
                    : 'text-gray-400 group-hover:text-gray-600',
                )}
              />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {/* Collapse Toggle */}
      <div className="border-t border-gray-200 px-3 py-2">
        <button
          onClick={onToggle}
          className="flex w-full items-center justify-center rounded-lg p-2 text-gray-400 hover:bg-gray-50 hover:text-gray-600 transition-colors"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? (
            <ChevronRight className="h-5 w-5" />
          ) : (
            <ChevronLeft className="h-5 w-5" />
          )}
        </button>
      </div>

      {/* User Info */}
      <div className="border-t border-gray-200 px-3 py-3">
        <div
          className={cn(
            'flex items-center gap-3',
            collapsed && 'justify-center',
          )}
        >
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-indigo-100 text-xs font-semibold text-indigo-700">
            {initials}
          </div>
          {!collapsed && user && (
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-gray-900">
                {user.first_name} {user.last_name}
              </p>
              <p className="truncate text-xs text-gray-500">{roleName}</p>
            </div>
          )}
        </div>
      </div>
    </aside>
  );
}
