export interface VehiclePhoto {
  id: string
  vehicle_id: string
  file_path: string
  sort_order: number
}

export interface Vehicle {
  id: string
  owner_id: string
  type: 'car' | 'motorbike'
  make: string
  model: string
  year: number
  transmission: 'auto' | 'manual'
  seats: number
  price_per_day: number
  location: string
  description: string
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
