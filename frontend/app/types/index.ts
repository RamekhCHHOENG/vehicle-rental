export interface VehiclePhoto {
  id: string
  vehicle_id: string
  file_path: string
  sort_order: number
}

export type VehicleType = 'car' | 'motorbike'

export interface Province {
  id: string
  code: string
  name_en: string
  name_km: string
  sort_order: number
  active: boolean
}

export interface VehicleModelRef {
  id: string
  make_id: string
  name: string
  type: VehicleType
  active: boolean
}

export interface VehicleMake {
  id: string
  name: string
  sort_order: number
  active: boolean
  /**
   * Whether the make has anything of each type. Derived server-side, because
   * with a few thousand makes imported the models are fetched per make rather
   * than shipped with the rest of the metadata.
   */
  has_cars?: boolean
  has_motorbikes?: boolean
  /** Only present on the admin payload, which does attach them. */
  models?: VehicleModelRef[]
}

export interface Feature {
  id: string
  code: string
  name: string
  icon: string
  /** Empty means it suits both cars and motorbikes. */
  applies_to?: VehicleType
  sort_order: number
  active: boolean
}

export interface SelectOption {
  value: string
  label: string
}

/** Everything the listing form and the browse filters need, in one payload. */
export interface Metadata {
  provinces: Province[]
  makes: VehicleMake[]
  features: Feature[]
  vehicle_types: SelectOption[]
  transmissions: SelectOption[]
  seat_options: number[]
}

export interface Vehicle {
  id: string
  owner_id: string
  type: VehicleType
  /**
   * The reference rows are nullable because a listing created before the
   * metadata tables existed may have resisted backfill. Read them through
   * `vehicleName` and `provinceName` rather than directly.
   */
  make_id?: string
  make?: VehicleMake
  model_id?: string
  model?: VehicleModelRef
  province_id?: string
  province?: Province
  year: number
  transmission: 'auto' | 'manual'
  seats: number
  price_per_day: number
  description: string
  features: Feature[]
  status: 'pending' | 'approved' | 'rejected'
  rejection_reason?: string
  photos: VehiclePhoto[]
  created_at: string
  owner?: { id: string, full_name: string, email: string }
  /** Only present on admin listings: bookings that a take-down would disrupt. */
  active_bookings?: number
}

export interface AdminUser {
  id: string
  email: string
  full_name: string
  phone: string
  role: 'renter' | 'owner' | 'admin'
  created_at: string
  vehicle_count: number
  booking_count: number
}

export interface Booking {
  id: string
  vehicle_id: string
  vehicle?: Vehicle
  renter_id: string
  renter?: { id: string, full_name: string, email: string, phone: string }
  start_date: string
  end_date: string
  total_price: number
  status: 'requested' | 'confirmed' | 'rejected' | 'cancelled' | 'completed'
  review?: { id: string, rating: number, comment: string }
  created_at: string
}
