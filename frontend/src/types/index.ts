export type AuthorizationStatus =
  | "pending_verification"
  | "demo_verified"
  | "authorized"
  | "expired"
  | "suspended"
  | "unknown"
  | string;

export type User = {
  id: string;
  role: "recycler" | "kabadiwala";
  name: string;
  phone: string;
  email: string;
  companyName?: string;
  facilityName?: string;
  businessName?: string;
  businessType?: string;
  gstNumber?: string;
  participationAllowed: boolean;
  effectiveAuthorizationStatus: AuthorizationStatus;
  verificationStatus: string;
  acceptedMaterials?: string[];
  processingCapacityKg?: number;
  authorization: {
    registrationNumber?: string;
    authority?: string;
    validFrom?: string;
    validUntil?: string;
    status: AuthorizationStatus;
    documentReference?: string;
    isDemo?: boolean;
  };
  pickup: { available: boolean; serviceArea?: string[] };
  offeredRates?: Array<{
    materialCategory: string;
    materialSubCategory?: string;
    pricePerUnit: number;
    unit: string;
  }>;
  address: { street: string; city: string; state: string; pincode: string };
};

export type EWasteLot = {
  id: string;
  referenceId: string;
  collectorId: string;
  domain: string;
  material: {
    category: string;
    subCategory?: string;
    description?: string;
    condition?: string;
    sourceType?: string;
    attributes?: Record<string, unknown>;
    subType?: string;
    grade?: string;
  };
  quantity: {
    approximateWeight: number;
    availableWeight: number;
    unit: string;
    total?: number;
    available?: number;
  };
  valuation: {
    estimatedValue: number;
    estimatedPricePerUnit: number;
    marketRangeMin: number;
    marketRangeMax: number;
    estimationSource?: string;
  };
  images?: string[];
  imageReferences?: Array<{ reference: string; capturedAt?: string }>;
  collection: {
    collectedAt?: string;
    location: { type: string; coordinates: number[] };
    label?: string;
  };
  status: string;
  createdAt: string;
};

export type LotView = {
  lot: EWasteLot;
  collector: {
    id: string;
    businessName: string;
    operatingLocation?: string;
    profileStatus: string;
    preferredLanguage?: string;
    memberSince: string;
  };
  distanceKm?: number;
};
export type ListingView = LotView;

export type OfferView = {
  offer: {
    id: string;
    lotId: string;
    collectorId: string;
    recyclerId: string;
    approximateWeight: number;
    offeredPricePerUnit: number;
    offeredTotal: number;
    unit: string;
    pickupAvailable: boolean;
    message?: string;
    status: string;
    createdAt: string;
  };
  materialName: string;
  collectorName: string;
  estimatedPricePerUnit: number;
  unit: string;
};

export type PriceBoard = {
  currency: string;
  notice: string;
  entries: Array<{
    materialCategory: string;
    materialSubCategory?: string;
    currentBuyingRate: number;
    quotedPrice: number;
    unit: string;
    city: string;
    state: string;
    marketRangeMin: number;
    marketRangeMax: number;
    trend: "up" | "down" | "stable" | string;
    updatedAt: string;
    sourceType: string;
    isDemo: boolean;
  }>;
};

export type Transaction = {
  id: string;
  reference: string;
  lotId: string;
  collectorId: string;
  recyclerId: string;
  materialSnapshot: {
    referenceId: string;
    material: EWasteLot["material"];
    sourceType?: string;
  };
  approximateQuantity: number;
  finalQuantity?: number;
  unit: string;
  quotedPricePerUnit: number;
  finalPricePerUnit?: number;
  quotedTotal: number;
  finalTotal?: number;
  collectionLabel?: string;
  handoverLabel?: string;
  status: string;
  paymentStatus: string;
  statusHistory: Array<{
    status: string;
    timestamp: string;
    actorType?: string;
    actorId?: string;
  }>;
  createdAt: string;
};

export type HandoverRecord = {
  id: string;
  handoverReference: string;
  lotId: string;
  transactionId: string;
  collectorId: string;
  recyclerId: string;
  materialLabel: string;
  photographs?: string[];
  approximateWeight: number;
  confirmedWeight?: number;
  weightUnit: string;
  timestamp: string;
  locationLabel?: string;
  collectorConfirmation: boolean;
  recyclerConfirmation: boolean;
  status: string;
  statusHistory: Array<{
    status: string;
    timestamp: string;
    actorType?: string;
  }>;
};

export type Payment = {
  id: string;
  paymentId: string;
  transactionId: string;
  recyclerId: string;
  collectorId: string;
  amount: number;
  currency: string;
  paymentMethod: "cash" | "digital" | string;
  status: string;
  paidAt?: string;
  createdAt: string;
};
