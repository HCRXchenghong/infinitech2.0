# 悦享e食用户端小程序参考图

把生成好的单页 UI 效果图放到这个目录，文件名与提示词文档保持一致。后续 UI 设计和前端还原都从这里找对应效果图：

- `00-welcome-legacy.png`
- `00-login-legacy.png`
- `00-register-legacy.png`
- `00-auth-login-register.png`
- `01-home.png`
- `02-shop-list.png`
- `03-shop-detail.png`
- `04-cart.png`
- `05-order-confirm.png`
- `06-order-list.png`
- `07-order-detail.png`
- `08-notification-preferences.png`
- `09-payment-password.png`
- `10-address-list.png`
- `11-circle.png`
- `12-meal-match.png`
- `13-after-sales.png`
- `14-order-review.png`
- `15-messages.png`
- `16-profile.png`
- `17-wallet.png`
- `18-coupons.png`
- `19-member-points.png`
- `20-invite-friends.png`
- `21-search.png`
- `22-medicine-home.png`
- `23-errand-home.png`
- `24-errand-order-detail.png`
- `25-merchant-group-chat.png`
- `26-red-packet-send.png`
- `27-red-packet-detail.png`
- `28-customer-service-chat.png`
- `29-service-ticket-detail.png`
- `30-complaint-feedback.png`
- `31-feedback-records.png`
- `32-prescription-upload.png`
- `33-prescription-review-result.png`
- `34-medicine-order-confirm.png`
- `35-medicine-order-detail.png`

后续前端还原时，以这些图片作为一比一视觉参考，落到 `apps/user-wechat-miniprogram/pages/**` 的 WXML/WXSS/TypeScript。

注意：`09-payment-password.png` 表达的是支付密码设置流程的参考状态，不是单屏同时承载全部逻辑。实际实现时「首次输入支付密码」和「再次确认支付密码」是两个页面/状态；安全提示是弹窗，不是页面常驻内容。
