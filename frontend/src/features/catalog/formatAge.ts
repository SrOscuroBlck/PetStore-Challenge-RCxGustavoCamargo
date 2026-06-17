/** Human age label: `1 → "1 yr"`, everything else `"n yrs"`. */
export function formatAge(ageYears: number): string {
  return ageYears === 1 ? '1 yr' : `${ageYears} yrs`;
}
