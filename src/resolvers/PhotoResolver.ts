import { Arg, Mutation, Query, Resolver } from 'type-graphql';
import { FileUpload, GraphQLUpload } from 'graphql-upload';
import Thumbnail, { ThumbnailType } from '../entities/Thumbnail';
import { bucketName, s3 } from '../helpers/s3';
import { createReadStream, createWriteStream } from 'fs';
import Photo from '../entities/photo';
import { getRepository } from 'typeorm';
import { join } from 'path';
import sharp from 'sharp';
import { stat } from 'fs/promises';
import { tmpdir } from 'os';
import { v4 } from 'uuid';

@Resolver(() => Photo)
export default class PhotoResolver {
  @Query(() => [Photo])
  public async photos(): Promise<Photo[]> {
    return Photo.getPhotos();
  }

  @Mutation(() => Photo)
  public async addPhoto(
    @Arg('photo', () => GraphQLUpload) file: FileUpload
  ): Promise<Photo> {
    // save photo to temp
    const uploadPath = await new Promise<string>((res, rej) => {
      const dir = join(tmpdir(), v4());
      file
        .createReadStream()
        .pipe(createWriteStream(dir))
        .on('finish', () => res(dir))
        .on('error', err => rej(err));
    });
    const [meta, stats] = await Promise.all([
      sharp(uploadPath).metadata(),
      stat(uploadPath),
    ]);
    if (!meta.width || !meta.height) {
      throw new Error('Cannot parse image');
    }
    const id = v4();
    const photo = new Photo(id);
    photo.size = stats.size;
    photo.width = meta.width;
    photo.height = meta.height;
    photo.mime = file.mimetype;
    await Promise.all([
      getRepository(Photo).save(photo),
      s3
        .putObject({
          ACL: 'private',
          Body: createReadStream(uploadPath),
          Bucket: bucketName,
          ContentLength: meta.size,
          ContentType: file.mimetype,
          Key: `photos/${id}`,
        })
        .promise(),
    ]);

    // generate thumbnails
    await Promise.all(
      new Array(2).fill(null).map(async (e, i) => {
        const thumbnailId = v4();
        const thumbnailPath = join(tmpdir(), v4());
        const image = sharp(uploadPath).resize({
          fit: 'inside',
          height: 500,
          width: 500,
        });

        const thumbnail =
          i === 1
            ? await image.blur(10).toFile(thumbnailPath)
            : await image.toFile(thumbnailPath);

        const thumbnailEntity = new Thumbnail(thumbnailId);
        thumbnailEntity.photo = Promise.resolve(photo);
        thumbnailEntity.height = thumbnail.height;
        thumbnailEntity.width = thumbnail.width;
        thumbnailEntity.size = thumbnail.size;
        thumbnailEntity.type =
          i === 1 ? ThumbnailType.BLUR : ThumbnailType.NORMAL;
        thumbnailEntity.mime = `image/${thumbnail.format}`;
        await Promise.all([
          getRepository(Thumbnail).save(thumbnailEntity),
          s3
            .putObject({
              ACL: 'private',
              Body: createReadStream(thumbnailPath),
              Bucket: bucketName,
              ContentLength: thumbnail.size,
              ContentType: 'image/png',
              Key: `thumbnails/${thumbnailId}`,
            })
            .promise(),
        ]);
        return thumbnailEntity;
      })
    );

    return photo;
  }
}
