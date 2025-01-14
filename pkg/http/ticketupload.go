package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/repository"
	"github.com/jadevelopmentgrp/Tickets-Archiver/pkg/repository/model"
	"go.uber.org/zap"
)

func (s *Server) ticketUploadHandler(ctx *gin.Context) {
	body, err := ctx.GetRawData()
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}

	guild, err := strconv.ParseUint(ctx.Query("guild"), 10, 64)
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing guild ID",
		})
		return
	}

	id, err := strconv.Atoi(ctx.Query("id"))
	if err != nil {
		ctx.JSON(400, gin.H{
			"message": "missing ticket ID",
		})
		return
	}

	// Get active bucket
	client, bucket, err := s.getActiveClient(ctx)
	if err != nil {
		s.Logger.Error("Failed to get active client", zap.Error(err))

		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Handle failed uploads: delete object from DB and S3 and re-upload
	var existingObject *model.Object
	if err := s.store.Tx(ctx, func(r repository.Repositories) error {
		obj, alreadyExists, err := r.Objects().GetObject(ctx, guild, id)
		if err != nil {
			return err
		}

		if alreadyExists {
			existingObject = &obj
		}

		return nil
	}); err != nil {
		s.Logger.Error("Failed to check whether object already exists", zap.Error(err))
		return
	}

	// If the object is already found, delete from S3
	if existingObject != nil {
		c, err := s.s3Clients.Get(existingObject.BucketId)
		if err != nil {
			s.Logger.Error("Failed to get client for existing object", zap.Error(err))
			ctx.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		if err := c.DeleteTicket(ctx, guild, id); err != nil {
			s.Logger.Error("Failed to delete existing ticket", zap.Error(err))
			ctx.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		if err := s.store.Tx(ctx, func(r repository.Repositories) error {
			return r.Objects().DeleteObject(ctx, guild, id)
		}); err != nil {
			s.Logger.Error("Failed to delete object from DB", zap.Error(err))
			ctx.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}
	}

	// Create object and commit transaction BEFORE writing to S3, to prevent "lost" objects
	if err := s.store.Tx(ctx, func(r repository.Repositories) error {
		return r.Objects().CreateObject(ctx, model.Object{
			GuildId:  guild,
			TicketId: id,
			BucketId: bucket.Id,
		})
	}); err != nil {
		s.Logger.Error("Failed to create object in DB", zap.Error(err))
		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := client.StoreTicket(ctx, guild, id, body); err != nil {
		s.Logger.Error("Failed to store ticket", zap.Error(err))

		// Try to remove object from DB, not the end of the world if it fails
		if err := s.store.Tx(ctx, func(r repository.Repositories) error {
			return r.Objects().DeleteObject(ctx, guild, id)
		}); err != nil {
			s.Logger.Error("Failed to delete object from DB", zap.Error(err))
		}

		ctx.JSON(500, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(200, gin.H{})
}
