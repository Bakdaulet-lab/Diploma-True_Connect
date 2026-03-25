// --- Auth ---
export interface AuthResponse {
  access_token: string;
  expires_in: number;
  user_id: string;
  refresh_token?: string;
}

export interface LoginRequest {
  phone: string;
  password: string;
}

export interface RegisterRequest {
  phone: string;
  password: string;
}

// --- User ---
export interface User {
  id: string;
  phone: string;
  email?: string;
  verification_level: 'none' | 'basic' | 'verified';
  trust_status?: string;
  trust_score: number;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
  updated_at?: string;
}

// --- Profile ---
export interface Profile {
  user_id: string;
  display_name: string;
  bio?: string;
  gender?: 'male' | 'female' | 'other';
  birth_date?: string;
  city?: string;
  latitude?: number;
  longitude?: number;
  looking_for?: string;
  avatar_url?: string;
  photos?: ProfilePhoto[];
  created_at?: string;
  updated_at?: string;
}

export interface ProfilePhoto {
  id: string;
  user_id: string;
  url: string;
  media_type: string;
  sort_order: number;
  is_verified: boolean;
  created_at: string;
}

export interface ProfileUpsert {
  display_name: string;
  bio?: string;
  gender?: string;
  birth_date?: string;
  city?: string;
  latitude?: number;
  longitude?: number;
  looking_for?: string;
}

// --- Match ---
export interface Match {
  id: string;
  user_a_id: string;
  user_b_id: string;
  user_a_liked: boolean;
  user_b_liked: boolean;
  matched_at?: string;
  created_at: string;
  other_user?: Profile;
  last_message?: Message;
}

// --- Message ---
export interface Message {
  id: string;
  match_id: string;
  sender_id: string;
  content: string;
  read_at?: string;
  created_at: string;
}

export interface SendMessagePayload {
  type: 'message';
  match_id: string;
  content: string;
}

export interface TypingPayload {
  type: 'typing';
  match_id: string;
}

// --- Post ---
export interface Post {
  id: string;
  author_id: string;
  content: string;
  media_url?: string;
  like_count: number;
  comment_count: number;
  is_liked: boolean;
  author_name?: string;     // <-- НОВОЕ ПОЛЕ
  author_avatar?: string;   // <-- НОВОЕ ПОЛЕ
  author?: Profile;
  created_at: string;
  updated_at?: string;
}

export interface PostComment {
  id: string;
  post_id: string;
  author_id: string;
  content: string;
  author_name?: string;     // <-- НОВОЕ ПОЛЕ
  author_avatar?: string;   // <-- НОВОЕ ПОЛЕ
  author?: Profile;
  created_at: string;
}
// --- Interaction / Trust ---
export interface Interaction {
  id: string;
  rater_id: string;
  rated_id: string;
  rating: number;
  context?: string;
  comment?: string;
  is_verified: boolean;
  created_at: string;
}

export interface TrustScore {
  user_id: string;
  score: number;
  rating_count: number;
}

// --- Settings ---
export interface UserSettings {
  user_id: string;
  push_notifications: boolean;
  show_online_status: boolean;
  distance_unit: string;
  max_distance_km: number;
  age_range_min: number;
  age_range_max: number;
  updated_at?: string;
}

// --- KYC ---
export interface KycStatus {
  status: 'none' | 'pending' | 'approved' | 'rejected';
  submitted_at?: string;
  reviewed_at?: string;
}

// --- Pagination ---
export interface PaginatedResponse<T> {
  items: T[];
  total?: number;
  page?: number;     // legacy offset pagination
  page_size?: number; // legacy offset pagination
  has_more?: boolean; // legacy offset fallback
  next_cursor?: string; // modern cursor pagination
  limit?: number;     // modern cursor
}

// --- Match Candidate ---
export type MatchCandidate = Profile;
