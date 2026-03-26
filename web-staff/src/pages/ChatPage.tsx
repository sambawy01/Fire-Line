import { useState, useEffect, useRef, useCallback } from 'react';
import {
  ArrowLeft,
  MapPin,
  Megaphone,
  MessageCircle,
  Send,
  Pin,
  Loader2,
} from 'lucide-react';
import { api } from '../lib/api';
import { getUser } from '../stores/auth';

/* ---------- types ---------- */

interface Channel {
  channel_id: string;
  org_id: string;
  location_id: string | null;
  name: string;
  type: 'location' | 'role' | 'direct' | 'broadcast';
  created_at: string;
}

interface Message {
  message_id: string;
  org_id: string;
  channel_id: string;
  sender_id: string;
  sender_name: string;
  sender_role: string;
  body: string;
  pinned: boolean;
  created_at: string;
}

/* ---------- helpers ---------- */

function formatTime(iso: string): string {
  const d = new Date(iso);
  const now = new Date();
  const diffMs = now.getTime() - d.getTime();
  const diffHours = diffMs / (1000 * 60 * 60);

  if (diffHours < 24) {
    return d.toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
  }
  if (diffHours < 168) {
    return d.toLocaleDateString('en-US', { weekday: 'short', hour: 'numeric', minute: '2-digit' });
  }
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
}

function channelIcon(type: string) {
  switch (type) {
    case 'location':
      return <MapPin size={18} className="text-blue-400" />;
    case 'broadcast':
      return <Megaphone size={18} className="text-orange-400" />;
    case 'direct':
      return <MessageCircle size={18} className="text-green-400" />;
    default:
      return <MessageCircle size={18} className="text-slate-400" />;
  }
}

const roleBadgeColors: Record<string, { bg: string; text: string }> = {
  owner: { bg: 'bg-purple-500/20', text: 'text-purple-400' },
  ops_director: { bg: 'bg-indigo-500/20', text: 'text-indigo-400' },
  gm: { bg: 'bg-blue-500/20', text: 'text-blue-400' },
  shift_manager: { bg: 'bg-orange-500/20', text: 'text-orange-400' },
  staff: { bg: 'bg-slate-500/20', text: 'text-slate-400' },
};

function roleBadge(role: string) {
  const colors = roleBadgeColors[role] || roleBadgeColors.staff;
  return (
    <span className={`inline-block px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-wider rounded ${colors.bg} ${colors.text}`}>
      {role.replace('_', ' ')}
    </span>
  );
}

/* ---------- Channel List ---------- */

function ChannelList({
  channels,
  lastMessages,
  loading,
  onSelect,
}: {
  channels: Channel[];
  lastMessages: Record<string, Message | undefined>;
  loading: boolean;
  onSelect: (ch: Channel) => void;
}) {
  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 size={28} className="animate-spin text-orange-400" />
      </div>
    );
  }

  if (channels.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-slate-500">
        <MessageCircle size={40} className="mb-3 opacity-40" />
        <p className="text-sm">No channels available</p>
      </div>
    );
  }

  return (
    <div className="divide-y divide-slate-700/50">
      {channels.map((ch) => {
        const last = lastMessages[ch.channel_id];
        return (
          <button
            key={ch.channel_id}
            onClick={() => onSelect(ch)}
            className="w-full flex items-start gap-3 px-4 py-3.5 text-left hover:bg-slate-800/60 active:bg-slate-800 transition-colors"
          >
            <div className="mt-0.5 shrink-0">{channelIcon(ch.type)}</div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between gap-2">
                <p className="text-sm font-semibold text-white truncate">{ch.name}</p>
                {last && (
                  <span className="text-[10px] text-slate-500 whitespace-nowrap">
                    {formatTime(last.created_at)}
                  </span>
                )}
              </div>
              {last ? (
                <p className="text-xs text-slate-400 truncate mt-0.5">
                  <span className="font-medium text-slate-300">{last.sender_name}:</span>{' '}
                  {last.body}
                </p>
              ) : (
                <p className="text-xs text-slate-500 italic mt-0.5">No messages yet</p>
              )}
            </div>
          </button>
        );
      })}
    </div>
  );
}

/* ---------- Chat View ---------- */

function ChatView({
  channel,
  onBack,
}: {
  channel: Channel;
  onBack: () => void;
}) {
  const user = getUser();
  const canPin = user && ['shift_manager', 'gm', 'ops_director', 'owner'].includes(user.role);

  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [draft, setDraft] = useState('');
  const [pinningId, setPinningId] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const fetchMessages = useCallback(async () => {
    try {
      const res = await api<{ messages: Message[] }>(
        `/messaging/channels/${channel.channel_id}/messages?limit=50`
      );
      // API returns newest first; reverse for display (oldest at top)
      setMessages((res.messages || []).reverse());
    } catch (err) {
      console.error('Failed to load messages', err);
    } finally {
      setLoading(false);
    }
  }, [channel.channel_id]);

  useEffect(() => {
    setLoading(true);
    setMessages([]);
    fetchMessages();

    // Poll for new messages every 5 seconds for real-time feel
    const interval = setInterval(fetchMessages, 5000);
    return () => clearInterval(interval);
  }, [fetchMessages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  async function handleSend() {
    if (!draft.trim() || !user || sending) return;

    setSending(true);
    try {
      const msg = await api<Message>(
        `/messaging/channels/${channel.channel_id}/messages`,
        {
          method: 'POST',
          body: JSON.stringify({
            body: draft.trim(),
            sender_name: user.display_name,
            sender_role: user.role,
          }),
        }
      );
      setMessages((prev) => [...prev, msg]);
      setDraft('');
      inputRef.current?.focus();
    } catch (err) {
      console.error('Failed to send message', err);
    } finally {
      setSending(false);
    }
  }

  async function handlePin(msg: Message) {
    if (!canPin || pinningId) return;
    setPinningId(msg.message_id);
    try {
      await api(`/messaging/messages/${msg.message_id}/pin`, {
        method: 'PUT',
        body: JSON.stringify({ pinned: !msg.pinned }),
      });
      setMessages((prev) =>
        prev.map((m) =>
          m.message_id === msg.message_id ? { ...m, pinned: !m.pinned } : m
        )
      );
    } catch (err) {
      console.error('Failed to pin message', err);
    } finally {
      setPinningId(null);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  const isOwn = (msg: Message) => user && msg.sender_id === user.user_id;

  return (
    <div className="flex flex-col h-full">
      {/* Header */}
      <div className="flex items-center gap-3 px-4 py-3 bg-slate-800 border-b border-slate-700 shrink-0">
        <button
          onClick={onBack}
          className="p-1 -ml-1 rounded-lg hover:bg-slate-700 transition-colors"
          aria-label="Back to channels"
        >
          <ArrowLeft size={20} className="text-slate-300" />
        </button>
        <div className="flex items-center gap-2">
          {channelIcon(channel.type)}
          <h2 className="text-sm font-semibold text-white">{channel.name}</h2>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-3">
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <Loader2 size={24} className="animate-spin text-orange-400" />
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-48 text-slate-500">
            <MessageCircle size={36} className="mb-2 opacity-40" />
            <p className="text-sm">No messages yet. Start the conversation!</p>
          </div>
        ) : (
          messages.map((msg) => {
            const own = isOwn(msg);
            return (
              <div
                key={msg.message_id}
                className={`flex ${own ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`relative max-w-[85%] rounded-2xl px-3.5 py-2.5 ${
                    own
                      ? 'bg-orange-500/20 border border-orange-500/10'
                      : 'bg-white/5 border border-white/5'
                  } ${msg.pinned ? 'ring-1 ring-amber-500/30' : ''}`}
                  onClick={() => canPin && handlePin(msg)}
                  role={canPin ? 'button' : undefined}
                  tabIndex={canPin ? 0 : undefined}
                >
                  {/* Pinned indicator */}
                  {msg.pinned && (
                    <div className="flex items-center gap-1 mb-1">
                      <Pin size={10} className="text-amber-400" />
                      <span className="text-[9px] font-medium text-amber-400 uppercase tracking-wider">
                        Pinned
                      </span>
                    </div>
                  )}

                  {/* Sender info (not shown for own messages) */}
                  {!own && (
                    <div className="flex items-center gap-1.5 mb-1">
                      <span className="text-xs font-bold text-white">
                        {msg.sender_name}
                      </span>
                      {roleBadge(msg.sender_role)}
                    </div>
                  )}

                  {/* Message body */}
                  <p className="text-[13px] leading-relaxed text-slate-200 whitespace-pre-wrap break-words">
                    {msg.body}
                  </p>

                  {/* Timestamp */}
                  <p
                    className={`text-[10px] mt-1 ${
                      own ? 'text-orange-400/60' : 'text-slate-500'
                    }`}
                  >
                    {formatTime(msg.created_at)}
                  </p>

                  {/* Pin loading indicator */}
                  {pinningId === msg.message_id && (
                    <div className="absolute top-1 right-1">
                      <Loader2 size={12} className="animate-spin text-amber-400" />
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
        <div ref={bottomRef} />
      </div>

      {/* Input bar */}
      <div className="shrink-0 px-3 py-2.5 bg-slate-800 border-t border-slate-700 safe-area-pb">
        <div className="flex items-center gap-2">
          <input
            ref={inputRef}
            type="text"
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            className="flex-1 px-4 py-2.5 text-sm bg-slate-700/60 border border-slate-600 rounded-full text-white placeholder-slate-400 outline-none focus:border-orange-500/50 focus:ring-1 focus:ring-orange-500/30 transition-all"
            disabled={sending}
          />
          <button
            onClick={handleSend}
            disabled={!draft.trim() || sending}
            className="flex items-center justify-center w-10 h-10 rounded-full bg-orange-500 text-white disabled:opacity-30 disabled:cursor-not-allowed hover:bg-orange-400 active:bg-orange-600 transition-all shrink-0"
            aria-label="Send message"
          >
            {sending ? (
              <Loader2 size={18} className="animate-spin" />
            ) : (
              <Send size={18} />
            )}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ---------- Main Page ---------- */

export default function ChatPage() {
  const user = getUser();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [lastMessages, setLastMessages] = useState<Record<string, Message | undefined>>({});
  const [loading, setLoading] = useState(true);
  const [selectedChannel, setSelectedChannel] = useState<Channel | null>(null);

  useEffect(() => {
    async function load() {
      try {
        const locationParam = user?.location_id ? `?location_id=${user.location_id}` : '';
        const res = await api<{ channels: Channel[] }>(`/messaging/channels${locationParam}`);
        const chList = res.channels || [];
        setChannels(chList);

        // Fetch last message for each channel
        const msgMap: Record<string, Message | undefined> = {};
        await Promise.all(
          chList.map(async (ch) => {
            try {
              const msgRes = await api<{ messages: Message[] }>(
                `/messaging/channels/${ch.channel_id}/messages?limit=1`
              );
              msgMap[ch.channel_id] = (msgRes.messages || [])[0];
            } catch {
              // ignore per-channel failures
            }
          })
        );
        setLastMessages(msgMap);
      } catch (err) {
        console.error('Failed to load channels', err);
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [user?.location_id]);

  // When a channel is selected, show the chat view fullscreen-style
  if (selectedChannel) {
    return (
      <div className="fixed inset-0 z-50 bg-slate-900 flex flex-col">
        <ChatView
          channel={selectedChannel}
          onBack={() => setSelectedChannel(null)}
        />
      </div>
    );
  }

  return (
    <div className="pb-4">
      {/* Page header */}
      <div className="px-4 pt-4 pb-3">
        <h1 className="text-lg font-bold text-white">Team Chat</h1>
        <p className="text-xs text-slate-400 mt-0.5">
          Stay connected with your team
        </p>
      </div>

      {/* Channel list */}
      <ChannelList
        channels={channels}
        lastMessages={lastMessages}
        loading={loading}
        onSelect={setSelectedChannel}
      />
    </div>
  );
}
