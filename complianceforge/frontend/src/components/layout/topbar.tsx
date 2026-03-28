'use client';

import { Menu, Bell } from 'lucide-react';
import { Breadcrumbs } from './breadcrumbs';
import { UserMenu } from './user-menu';

// ── Types ────────────────────────────────────────────────
interface TopbarUser {
  first_name: string;
  last_name: string;
  email: string;
}

interface TopbarProps {
  onMenuToggle: () => void;
  user?: TopbarUser;
}

// ── Component ────────────────────────────────────────────
export function Topbar({ onMenuToggle, user }: TopbarProps) {
  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4 sm:px-6">
      {/* Left: hamburger + breadcrumbs */}
      <div className="flex items-center gap-4">
        <button
          onClick={onMenuToggle}
          className="rounded-lg p-2 text-gray-400 hover:bg-gray-50 hover:text-gray-600 lg:hidden"
          aria-label="Toggle menu"
        >
          <Menu className="h-5 w-5" />
        </button>

        <div className="hidden sm:block">
          <Breadcrumbs />
        </div>
      </div>

      {/* Right: notifications + user menu */}
      <div className="flex items-center gap-3">
        <button
          className="relative rounded-lg p-2 text-gray-400 hover:bg-gray-50 hover:text-gray-600 transition-colors"
          aria-label="Notifications"
        >
          <Bell className="h-5 w-5" />
          {/* Notification dot */}
          <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500" />
        </button>

        {user && <UserMenu user={user} />}
      </div>
    </header>
  );
}
