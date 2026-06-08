import { blockMealMatchCandidate, ensurePreviewAuth, getMealMatchCandidates, getMealMatchProfile, reportMealMatchCandidate, saveMealMatchProfile } from "../../utils/api";

function getMealMatchDeviceId() {
  const storageKey = "mealMatchDeviceId";
  const existing = String(wx.getStorageSync(storageKey) || "");
  if (existing) return existing;
  const created = `wx_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`;
  wx.setStorageSync(storageKey, created);
  return created;
}

Page({
  data: {
    status: "未开启",
    privacyPolicy: "同校可见 · 隐藏楼栋和精确位置",
    deviceRiskText: "设备环境将在提交资料时校验",
    requirements: [
      { title: "完善性别与基础资料", status: "待完成", done: false },
      { title: "确认同校/同楼隐私范围", status: "待确认", done: false },
      { title: "完成设备安全校验", status: "待校验", done: false },
      { title: "签署身份真实性承诺", status: "待签署", done: false },
      { title: "签署平台免责承诺", status: "待签署", done: false },
      { title: "完成性格与饮食习惯问卷", status: "未完成", done: false },
      { title: "通过资料人工审核", status: "待提交", done: false }
    ],
    emptyText: "完成资料并通过平台审核后，系统会展示推荐饭搭。",
    candidates: [
      { id: "user_buddy_lunch", title: "同楼午餐搭子", subtitle: "偏好：简餐、咖啡", distance: "同楼可约", location: "无限科技大学 · 东区食堂", scope: "同楼可见", score: "待开启", privacy: "完成资料后展示匹配标签" },
      { id: "user_buddy_weekend", title: "周末探店搭子", subtitle: "偏好：火锅、烤肉", distance: "同校范围", location: "无限科技大学 · 东区", scope: "同校可见", score: "待开启", privacy: "不公开手机号与精确位置" }
    ]
  },
  onShow() {
    this.loadProfile();
  },
  async loadProfile() {
    ensurePreviewAuth();
    try {
      const result = await getMealMatchProfile() as { profile?: Record<string, unknown>; can_use?: boolean; missing?: string[] };
      this.applyProfile(result.profile || {}, Boolean(result.can_use), result.missing || []);
      await this.loadCandidates();
    } catch (_error) {
      // Keep the initial locked state in pure preview mode.
    }
  },
  async loadCandidates() {
    try {
      const result = await getMealMatchCandidates() as { candidates?: Record<string, unknown>[]; missing?: string[]; review_required?: boolean; privacy_notice?: string; privacy_scope?: string; device_risk_state?: string; device_risk_reason?: string };
      const candidates = (result.candidates || []).map(formatCandidate);
      this.setData({
        candidates,
        privacyPolicy: String(result.privacy_notice || privacyText(result.privacy_scope)),
        deviceRiskText: deviceRiskText(result.device_risk_state, result.device_risk_reason),
        emptyText: candidates.length > 0 ? "附近暂时没有新的推荐饭搭，稍后再来看看。" : emptyTextForMissing(result.missing || [], Boolean(result.review_required))
      });
    } catch (_error) {
      // Keep seeded candidate cards when the API service is not running.
    }
  },
  async handleStartTap() {
    ensurePreviewAuth();
    try {
      const result = await saveMealMatchProfile({
        gender: "undisclosed",
        school_id: "infinitech_university",
        school_name: "无限科技大学",
        campus_name: "东区",
        building_id: "east_canteen",
        building_name: "东区食堂",
        privacy_scope: "same_building",
        location_precision: "building_only",
        device_id: getMealMatchDeviceId(),
        identity_truth_signed: true,
        platform_liability_release_signed: true,
        questionnaire_completed: true,
        personality_traits: ["细心", "守时"],
        dietary_habits: ["清淡", "不浪费"]
      }) as { profile?: Record<string, unknown>; can_use?: boolean; missing?: string[] };
      this.applyProfile(result.profile || {}, Boolean(result.can_use), result.missing || []);
      await this.loadCandidates();
      wx.showToast({ title: result.can_use ? "资料已完善" : "已提交审核", icon: "success" });
    } catch (_error) {
      this.setData({
        status: "预览可开启",
        requirements: this.data.requirements.map((item) => ({ ...item, status: "已完成", done: true }))
      });
      wx.showToast({ title: "预览资料已完善", icon: "none" });
    }
  },
  applyProfile(profile, canUse, missing = []) {
    const personalityTraits = profile.personality_traits || [];
    const dietaryHabits = profile.dietary_habits || [];
    const moderationStatus = String(profile.moderation_status || "");
    const moderationReason = String(profile.moderation_reason || "");
    const privacyScope = String(profile.privacy_scope || "same_school");
    const deviceRiskState = String(profile.device_risk_state || "");
    const deviceRiskReason = String(profile.device_risk_reason || "");
    const reviewDone = canUse || moderationStatus === "approved";
    const reviewPending = hasMissing(missing, "moderation_pending") || moderationStatus === "pending_review";
    const reviewRejected = hasMissing(missing, "moderation_rejected") || moderationStatus === "rejected";
    this.setData({
      status: canUse ? "可开启" : reviewPending ? "审核中" : reviewRejected ? "审核未通过" : "待完善",
      emptyText: canUse ? "附近暂时没有新的推荐饭搭，稍后再来看看。" : emptyTextForMissing(missing, reviewPending, moderationReason),
      privacyPolicy: privacyText(privacyScope),
      deviceRiskText: deviceRiskText(deviceRiskState, deviceRiskReason),
      requirements: [
        requirement("完善性别与基础资料", Boolean(profile.gender), "待完成"),
        requirement("确认同校/同楼隐私范围", Boolean(profile.school_id) && Boolean(profile.privacy_scope), "待确认"),
        requirement("完成设备安全校验", deviceRiskState === "passed", deviceRiskState === "review" ? "需复核" : "待校验"),
        requirement("签署身份真实性承诺", Boolean(profile.identity_truth_signed), "待签署"),
        requirement("签署平台免责承诺", Boolean(profile.platform_liability_release_signed), "待签署"),
        requirement("完成性格与饮食习惯问卷", Boolean(profile.questionnaire_completed) && personalityTraits.length > 0 && dietaryHabits.length > 0, "未完成"),
        requirement("通过资料人工审核", reviewDone, reviewRejected ? "需修改" : reviewPending ? "审核中" : "待提交")
      ]
    });
  },
  async handleBlockTap(event: WechatMiniprogram.BaseEvent) {
    const targetUserId = String(event.currentTarget.dataset.id || "");
    if (!targetUserId) return;
    ensurePreviewAuth();
    try {
      await blockMealMatchCandidate({ target_user_id: targetUserId, reason: "not_interested" });
      const candidates = this.data.candidates.filter((item) => item.id !== targetUserId);
      this.setData({ candidates, emptyText: candidates.length > 0 ? this.data.emptyText : "附近暂时没有新的推荐饭搭，稍后再来看看。" });
      wx.showToast({ title: "已减少推荐", icon: "none" });
    } catch (_error) {
      const candidates = this.data.candidates.filter((item) => item.id !== targetUserId);
      this.setData({ candidates, emptyText: candidates.length > 0 ? this.data.emptyText : "附近暂时没有新的推荐饭搭，稍后再来看看。" });
      wx.showToast({ title: "已在本地隐藏", icon: "none" });
    }
  },
  async handleReportTap(event: WechatMiniprogram.BaseEvent) {
    const targetUserId = String(event.currentTarget.dataset.id || "");
    if (!targetUserId) return;
    ensurePreviewAuth();
    try {
      await reportMealMatchCandidate({
        target_user_id: targetUserId,
        reason: "unsafe_or_fake_profile",
        description: "用户端快速举报"
      });
      wx.showToast({ title: "举报已提交", icon: "success" });
    } catch (_error) {
      wx.showToast({ title: "举报已记录", icon: "none" });
    }
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function requirement(title: string, done: boolean, pending: string) {
  return { title, done, status: done ? "已完成" : pending };
}

function hasMissing(missing: string[], key: string) {
  return missing.some((item) => item === key);
}

function emptyTextForMissing(missing: string[], reviewRequired: boolean, moderationReason = "") {
  if (hasMissing(missing, "device_risk_review")) {
    return "设备环境需要人工复核，审核通过后会展示推荐饭搭。";
  }
  if (hasMissing(missing, "moderation_rejected")) {
    return moderationReason || "资料审核未通过，修改后可重新提交。";
  }
  if (reviewRequired || hasMissing(missing, "moderation_pending")) {
    return "资料已提交，平台审核通过后会展示推荐饭搭。";
  }
  if (missing.length > 0) {
    return "完成资料、同校隐私范围、设备校验、承诺和问卷后，再开启推荐饭搭。";
  }
  return "附近暂时没有新的推荐饭搭，稍后再来看看。";
}

function formatCandidate(candidate: Record<string, unknown>) {
  const matchedDietary = Array.isArray(candidate.matched_dietary_habits) ? candidate.matched_dietary_habits : [];
  const matchedTraits = Array.isArray(candidate.matched_personality_traits) ? candidate.matched_personality_traits : [];
  const tags = [...matchedDietary, ...matchedTraits].map((item) => String(item)).filter(Boolean);
  const sameBuilding = Boolean(candidate.same_building);
  const schoolName = String(candidate.school_name || "同校");
  const campusName = String(candidate.campus_name || "校区");
  const buildingName = String(candidate.building_name || "");
  return {
    id: String(candidate.user_id || ""),
    title: String(candidate.display_name || "饭搭用户"),
    subtitle: tags.length ? `匹配：${tags.slice(0, 3).join("、")}` : "偏好已通过平台保护展示",
    distance: String(candidate.distance_text || "附近"),
    location: sameBuilding && buildingName ? `${schoolName} · ${buildingName}` : `${schoolName} · ${campusName}`,
    scope: sameBuilding ? "同楼可见" : "同校可见",
    score: `${Number(candidate.match_score || 0)}分`,
    privacy: String(candidate.privacy_notice || "不公开手机号与精确位置")
  };
}

function privacyText(scope?: unknown) {
  return String(scope || "") === "same_building" ? "同校同楼可见 · 隐藏精确位置" : "同校可见 · 隐藏楼栋和精确位置";
}

function deviceRiskText(state?: unknown, reason?: unknown) {
  if (String(state || "") === "passed") return "设备环境已通过";
  if (String(state || "") === "review") return String(reason || "设备环境需人工复核");
  if (String(state || "") === "blocked") return "设备环境高风险，暂不可开启";
  return "设备环境将在提交资料时校验";
}
