import {
  BeforeInsert,
  BeforeUpdate,
  Column,
  CreateDateColumn,
  Entity,
  ManyToOne,
  PrimaryColumn,
} from 'typeorm';
import {
  IsNotEmpty,
  IsUUID,
  Max,
  MaxLength,
  Min,
  validateOrReject,
} from 'class-validator';
import Photo from './photo';

export enum ThumbnailType {
  NORMAL = 'NORMAL',
  BLUR = 'BLUR',
}

@Entity({ name: 'thumbnails' })
export default class Thumbnail {
  @PrimaryColumn('uuid')
  @IsUUID('4')
  public readonly id: string;

  @Column({ type: 'int' })
  @IsNotEmpty()
  @Min(1)
  @Max(2147483647)
  public size: number;

  @Column({ type: 'smallint' })
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public width: number;

  @Column({ type: 'smallint' })
  @IsNotEmpty()
  @Min(1)
  @Max(32767)
  public height: number;

  @Column({ length: 32 })
  @IsNotEmpty()
  @MaxLength(32)
  public mime: string;

  @Column({ enum: ThumbnailType, type: 'enum' })
  public type: ThumbnailType;

  @ManyToOne(() => Photo, photo => photo.thumbnails, { onDelete: 'CASCADE' })
  public photo: Promise<Photo>;

  @CreateDateColumn()
  public readonly uploadedTime: Date;

  public constructor(id: string) {
    this.id = id;
  }

  @BeforeInsert()
  @BeforeUpdate()
  public async validate(): Promise<void> {
    await validateOrReject(this);
  }
}
