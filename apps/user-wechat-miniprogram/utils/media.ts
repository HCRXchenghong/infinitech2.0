export const generatedImages = {
  homeHero: "/assets/generated/home-hero.jpg",
  homeFeaturedDish: "/assets/generated/home-featured-dish.jpg",
  homeRecommendRestaurant: "/assets/generated/home-recommend-restaurant.jpg",
  homeRecommendCourier: "/assets/generated/home-recommend-courier.jpg",
  homeMealMatch: "/assets/generated/home-meal-match.jpg",
  shopCoffee: "/assets/generated/shop-coffee.jpg",
  shopPharmacy: "/assets/generated/shop-pharmacy.jpg",
  shopHomeCooking: "/assets/generated/shop-home-cooking.jpg",
  shopClaypot: "/assets/generated/shop-claypot.jpg",
  shopDetailCover: "/assets/generated/shop-detail-cover.jpg",
  productBeefRice: "/assets/generated/product-beef-rice.jpg",
  productLemonTea: "/assets/generated/product-lemon-tea.jpg",
  productChickenWings: "/assets/generated/product-chicken-wings.jpg",
  productTomatoNoodle: "/assets/generated/product-tomato-noodle.jpg",
  medicineClinicCover: "/assets/generated/medicine-clinic-cover.jpg",
  medicineCoolingPatch: "/assets/generated/medicine-cooling-patch.jpg",
  medicineCapsules: "/assets/generated/medicine-capsules.jpg",
  medicineFirstAid: "/assets/generated/medicine-first-aid.jpg",
  errandHero: "/assets/generated/errand-hero.jpg",
  errandParcel: "/assets/generated/errand-parcel.jpg",
  categoryTakeout: "/assets/generated/category-takeout.png",
  categoryGroupbuy: "/assets/generated/category-groupbuy.png",
  categoryMedicine: "/assets/generated/category-medicine.png",
  categoryCourier: "/assets/generated/category-courier.png",
  categoryCircle: "/assets/generated/category-circle.png",
  categoryMealMatch: "/assets/generated/category-meal-match.png",
  categoryCoupons: "/assets/generated/category-coupons.png",
  categoryPoints: "/assets/generated/category-points.png",
  tabbarHome: "/assets/generated/tabbar/tabbar-home.png",
  tabbarHomeActive: "/assets/generated/tabbar/tabbar-home-active.png",
  tabbarOrders: "/assets/generated/tabbar/tabbar-orders.png",
  tabbarOrdersActive: "/assets/generated/tabbar/tabbar-orders-active.png",
  tabbarMessages: "/assets/generated/tabbar/tabbar-messages.png",
  tabbarMessagesActive: "/assets/generated/tabbar/tabbar-messages-active.png",
  tabbarProfile: "/assets/generated/tabbar/tabbar-profile.png",
  tabbarProfileActive: "/assets/generated/tabbar/tabbar-profile-active.png"
};

export function mediaUrl(value: unknown, fallback = "") {
  const url = String(value || "").trim();
  return url || fallback;
}

export function shopFallbackImage(shopId: string, index = 0) {
  const map: Record<string, string> = {
    shop_1: generatedImages.homeRecommendRestaurant,
    shop_coffee_preview: generatedImages.shopCoffee,
    shop_medicine_preview: generatedImages.shopPharmacy,
    shop_kitchen_preview: generatedImages.shopHomeCooking,
    shop_pot_preview: generatedImages.shopClaypot
  };
  const sequence = [
    generatedImages.homeRecommendRestaurant,
    generatedImages.shopCoffee,
    generatedImages.shopPharmacy,
    generatedImages.shopHomeCooking,
    generatedImages.shopClaypot
  ];
  return map[shopId] || sequence[index % sequence.length] || generatedImages.homeRecommendRestaurant;
}

export function productFallbackImage(productId: string) {
  const map: Record<string, string> = {
    prod_beef_rice: generatedImages.productBeefRice,
    prod_tea: generatedImages.productLemonTea,
    prod_wings: generatedImages.productChickenWings,
    prod_noodle: generatedImages.productTomatoNoodle,
    prod_soup: generatedImages.productLemonTea
  };
  return map[productId] || generatedImages.productBeefRice;
}

export function medicineFallbackImage(productId: string) {
  const map: Record<string, string> = {
    med_cooling_patch: generatedImages.medicineCoolingPatch,
    med_amoxicillin: generatedImages.medicineCapsules,
    med_swab: generatedImages.medicineFirstAid,
    med_bandage: generatedImages.medicineFirstAid
  };
  return map[productId] || generatedImages.medicineFirstAid;
}

export function errandFallbackImage() {
  return generatedImages.errandParcel;
}

export function categoryIconImage(key: string) {
  const map: Record<string, string> = {
    takeout: generatedImages.categoryTakeout,
    groupbuy: generatedImages.categoryGroupbuy,
    medicine: generatedImages.categoryMedicine,
    courier: generatedImages.categoryCourier,
    circle: generatedImages.categoryCircle,
    "meal-match": generatedImages.categoryMealMatch,
    coupons: generatedImages.categoryCoupons,
    points: generatedImages.categoryPoints
  };
  return map[key] || generatedImages.categoryTakeout;
}

export function tabbarIconImage(key: string, active = false) {
  const map: Record<string, { normal: string; active: string }> = {
    home: { normal: generatedImages.tabbarHome, active: generatedImages.tabbarHomeActive },
    orders: { normal: generatedImages.tabbarOrders, active: generatedImages.tabbarOrdersActive },
    messages: { normal: generatedImages.tabbarMessages, active: generatedImages.tabbarMessagesActive },
    profile: { normal: generatedImages.tabbarProfile, active: generatedImages.tabbarProfileActive }
  };
  const item = map[key] || map.home;
  return active ? item.active : item.normal;
}
