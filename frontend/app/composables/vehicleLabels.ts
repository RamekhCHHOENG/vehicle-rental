import type { Vehicle } from '~/types'

// Make, model and province live on the vehicle as objects now, and they are
// nullable for listings that predate the reference tables. Every screen that
// names a vehicle goes through here, so there is one answer to "what if it is
// missing" instead of a different guess in each template.

/** "Toyota Camry" — or an honest placeholder when the references are missing. */
export function vehicleName(vehicle?: Partial<Vehicle> | null): string {
  if (!vehicle) return 'Vehicle'

  const parts = [vehicle.make?.name, vehicle.model?.name].filter(Boolean)
  return parts.length ? parts.join(' ') : 'Unspecified vehicle'
}

/** The province a vehicle is in, in English. */
export function provinceName(vehicle?: Partial<Vehicle> | null): string {
  return vehicle?.province?.name_en ?? 'Location not set'
}

/** Province in both scripts, for the detail page where there is room. */
export function provinceNameFull(vehicle?: Partial<Vehicle> | null): string {
  const province = vehicle?.province
  if (!province) return 'Location not set'
  return province.name_km ? `${province.name_en} · ${province.name_km}` : province.name_en
}
