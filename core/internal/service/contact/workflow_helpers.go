package contact

import (
	"context"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

func AddTagToContact(ctx context.Context, email string, groupId int, tagName string) error {
	var contactId int64
	if err := g.DB().Model("bm_contacts").Ctx(ctx).
		Fields("id").
		Where("email", email).
		Where("group_id", groupId).
		Scan(&contactId); err != nil {
		return fmt.Errorf("contact lookup: %w", err)
	}
	if contactId == 0 {
		return fmt.Errorf("contact %q not found in group %d", email, groupId)
	}

	var tagId int64
	if err := g.DB().Model("bm_tags").Ctx(ctx).
		Fields("id").
		Where("name", tagName).
		Where("group_id", groupId).
		Scan(&tagId); err != nil {
		return fmt.Errorf("tag lookup: %w", err)
	}
	if tagId == 0 {
		res, err := g.DB().Model("bm_tags").Ctx(ctx).
			Data(g.Map{"name": tagName, "group_id": groupId, "create_time": time.Now().Unix()}).
			InsertAndGetId()
		if err != nil {
			return fmt.Errorf("create tag: %w", err)
		}
		tagId = res
	}

	_, err := g.DB().Model("bm_contact_tags").Ctx(ctx).
		Data(g.Map{"contact_id": contactId, "tag_id": tagId, "create_time": time.Now().Unix()}).
		InsertIgnore()
	return err
}

func RemoveTagFromContact(ctx context.Context, email string, groupId int, tagName string) error {
	var contactId int64
	if err := g.DB().Model("bm_contacts").Ctx(ctx).
		Fields("id").
		Where("email", email).
		Where("group_id", groupId).
		Scan(&contactId); err != nil || contactId == 0 {
		return nil
	}

	var tagId int64
	if err := g.DB().Model("bm_tags").Ctx(ctx).
		Fields("id").
		Where("name", tagName).
		Where("group_id", groupId).
		Scan(&tagId); err != nil || tagId == 0 {
		return nil
	}

	_, err := g.DB().Model("bm_contact_tags").Ctx(ctx).
		Where("contact_id", contactId).
		Where("tag_id", tagId).
		Delete()
	return err
}

func MoveContactToGroup(ctx context.Context, email string, targetGroupId int) error {
	_, err := g.DB().Model("bm_contacts").Ctx(ctx).
		Data(g.Map{"group_id": targetGroupId}).
		Where("email", email).
		Update()
	return err
}
