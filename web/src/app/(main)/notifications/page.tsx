'use client';

import { useEffect, Fragment } from 'react';
import { useNotifications, useMarkAllNotificationsAsRead, useMarkNotificationAsRead } from '@/hooks/api';
import { AppNotification } from '@/types';
import { formatDistanceToNow } from 'date-fns';
import { Heart, MessageSquare, UserPlus, Bell, Check, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import Link from 'next/link';

export default function NotificationsPage() {
  const {
    data,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    status,
  } = useNotifications();

  const markAllAsRead = useMarkAllNotificationsAsRead();
  const markAsRead = useMarkNotificationAsRead();

  useEffect(() => {
    // Load more when scrolled to bottom
    const handleScroll = () => {
      if (
        window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 500 &&
        hasNextPage &&
        !isFetchingNextPage
      ) {
        fetchNextPage();
      }
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, [hasNextPage, isFetchingNextPage, fetchNextPage]);

  if (status === 'pending') {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary-500" />
      </div>
    );
  }

  if (status === 'error') {
    return (
      <div className="p-8 text-center text-gray-500 bg-white rounded-2xl shadow-sm border border-gray-100">
        Failed to load notifications.
      </div>
    );
  }

  const handleMarkAllRead = () => {
    markAllAsRead.mutate();
  };

  const pages = data?.pages || [];
  const isEmpty = pages.length === 0 || pages[0].items.length === 0;

  return (
    <div className="mx-auto max-w-2xl w-full">
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold tracking-tight text-gray-900">Notifications</h1>
        {!isEmpty && (
          <button
            onClick={handleMarkAllRead}
            disabled={markAllAsRead.isPending}
            className="flex items-center gap-2 rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition hover:bg-gray-200 disabled:opacity-50"
          >
            <Check className="h-4 w-4" />
            Mark all right as read
          </button>
        )}
      </div>

      {isEmpty ? (
        <div className="flex flex-col items-center justify-center rounded-2xl border border-gray-100 bg-white p-12 text-center shadow-sm">
          <div className="mb-4 rounded-full bg-gray-50 p-4">
            <Bell className="h-8 w-8 text-gray-400" />
          </div>
          <h3 className="text-lg font-semibold text-gray-900">No notifications yet</h3>
          <p className="text-gray-500 max-w-sm mt-2">
            When you get new matches, likes, or comments on your posts, they'll show up here.
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {pages.map((page, i) => (
            <Fragment key={i}>
              {page.items.map((notification) => (
                <NotificationItem
                  key={notification.id}
                  notification={notification}
                  onRead={() => {
                    if (!notification.is_read) {
                      markAsRead.mutate(notification.id);
                    }
                  }}
                />
              ))}
            </Fragment>
          ))}
          {isFetchingNextPage && (
            <div className="py-4 text-center">
              <Loader2 className="mx-auto h-6 w-6 animate-spin text-gray-400" />
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function NotificationItem({
  notification,
  onRead,
}: {
  notification: AppNotification;
  onRead: () => void;
}) {
  const isUnread = !notification.is_read;

  let icon = <Bell className="h-5 w-5 text-gray-500" />;
  let colorClass = 'bg-gray-100';
  let message = 'New notification';
  let link = '#';

  if (notification.type === 'like') {
    icon = <Heart className="h-5 w-5 text-red-500" />;
    colorClass = 'bg-red-50';
    message = 'liked your post.';
    if (notification.entity_id) {
      link = `/feed?post=\${notification.entity_id}`;
    }
  } else if (notification.type === 'comment') {
    icon = <MessageSquare className="h-5 w-5 text-blue-500" />;
    colorClass = 'bg-blue-50';
    message = 'commented on your post.';
    if (notification.entity_id) {
      link = `/feed?post=\${notification.entity_id}`;
    }
  } else if (notification.type === 'match') {
    icon = <UserPlus className="h-5 w-5 text-primary-500" />;
    colorClass = 'bg-primary-50';
    message = 'matched with you! Send them a message.';
    if (notification.entity_id) {
      link = `/matches/\${notification.entity_id}`;
    }
  }

  return (
    <Link
      href={link}
      onClick={() => onRead()}
      className={cn(
        'group flex items-start gap-4 rounded-2xl border p-4 transition-all hover:shadow-md',
        isUnread
          ? 'border-primary-100 bg-primary-50/30'
          : 'border-gray-100 bg-white hover:border-gray-200'
      )}
    >
      <div className={cn('flex flex-shrink-0 items-center justify-center rounded-full p-3', colorClass)}>
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm text-gray-900">
          {notification.actor_name && (
            <span className="font-bold mr-1">{notification.actor_name}</span>
          )}
          <span className="text-gray-600">{message}</span>
        </p>
        <span className="text-xs font-medium text-gray-400 mt-1 block">
          {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
        </span>
      </div>
      {isUnread && (
        <span className="flex h-2.5 w-2.5 flex-shrink-0 rounded-full bg-primary-500" />
      )}
    </Link>
  );
}
